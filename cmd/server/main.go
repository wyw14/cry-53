package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/wyw14/cry-053/internal/application"
	"github.com/wyw14/cry-053/internal/config"
	platformcrypto "github.com/wyw14/cry-053/internal/platform/crypto"
	"github.com/wyw14/cry-053/internal/platform/health"
	"github.com/wyw14/cry-053/internal/platform/identity"
	"github.com/wyw14/cry-053/internal/platform/notify"
	"github.com/wyw14/cry-053/internal/repository/memory"
	"github.com/wyw14/cry-053/internal/service"
	transport "github.com/wyw14/cry-053/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	settings, err := config.Load()
	if err != nil {
		logger.Fatal("load configuration", zap.Error(err))
	}
	store := memory.NewStore()
	types := memory.Types(store)
	configs := memory.Configurations(store)
	bundles := memory.Bundles(store)
	versions := memory.Versions(store)
	audits := memory.Audits(store)
	usages := memory.Usages(store)
	if err := service.SeedTypes(context.Background(), types); err != nil {
		logger.Fatal("seed config types", zap.Error(err))
	}
	cipher, err := platformcrypto.NewAESCipher(settings.MasterKey)
	if err != nil {
		logger.Fatal("initialize cipher", zap.Error(err))
	}
	_ = cipher
	ids := &identity.Generator{}
	notifier := &notify.MemoryNotifier{}
	parser := service.NewBundleParser(settings.MaxUploadBytes)
	schema := service.NewSchemaValidator(types)
	dependencies := service.NewDependencyValidator(types, configs)
	diff := service.NewDiffService(configs, types)
	workflow := application.NewBundleWorkflow(parser, schema, dependencies, diff, bundles, memory.NewManager(store), application.SystemClock{}, ids, notifier)
	catalog := application.NewTypeCatalog(types)
	lifecycle := application.NewLifecycleService(configs, versions, audits, application.SystemClock{}, ids)
	impact := application.NewImpactAnalyzer(configs, usages, health.NewEndpointAdapter(settings.RequestTimeout))
	auditTrail := application.NewAuditTrail(audits)
	handler := transport.NewHandler(workflow, catalog, lifecycle, impact, auditTrail, settings.MaxUploadBytes)
	readiness := &transport.Readiness{}
	router := transport.NewRouter(handler, logger, settings.RequestTimeout, readiness)
	server := &http.Server{Addr: settings.HTTPAddr, Handler: router, ReadHeaderTimeout: settings.RequestTimeout, IdleTimeout: settings.RequestTimeout * 12}
	readiness.Set(true)
	go func() {
		logger.Info("server started", zap.String("address", settings.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("serve", zap.Error(err))
		}
	}()
	stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-stop.Done()
	readiness.Set(false)
	shutdown, shutdownCancel := context.WithTimeout(context.Background(), settings.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdown); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}
}

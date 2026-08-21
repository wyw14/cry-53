package http

import (
	"bytes"
	"io"
	stdhttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-053/internal/application"
	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/middleware"
	"github.com/wyw14/cry-053/internal/service"
)

type Handler struct {
	bundles   *application.BundleWorkflow
	types     *application.TypeCatalog
	lifecycle *application.LifecycleService
	impact    *application.ImpactAnalyzer
	audits    *application.AuditTrail
	maxUpload int64
}

type operation[T any] func() (T, error)

func render[T any](c *gin.Context, status int, run operation[T]) {
	result, err := run()
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(status, result)
}

func decode[T any](c *gin.Context) (T, bool) {
	var request T
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, &domain.ValidationError{Code: "INVALID_BODY", Fields: []domain.FieldError{{Path: "$", Code: "json", Message: "请求体不是合法 JSON"}}})
		return request, false
	}
	return request, true
}

func NewHandler(bundles *application.BundleWorkflow, types *application.TypeCatalog, lifecycle *application.LifecycleService, impact *application.ImpactAnalyzer, audits *application.AuditTrail, maxUpload int64) *Handler {
	return &Handler{bundles: bundles, types: types, lifecycle: lifecycle, impact: impact, audits: audits, maxUpload: maxUpload}
}

func (h *Handler) uploadBundle(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		writeError(c, &domain.ValidationError{Code: "FILE_REQUIRED", Fields: []domain.FieldError{{Path: "file", Code: "required", Message: "请选择 JSON 配置包"}}})
		return
	}
	defer file.Close()
	if header.Size > h.maxUpload {
		writeError(c, &domain.ValidationError{Code: "FILE_TOO_LARGE", Fields: []domain.FieldError{{Path: "file", Code: "max_bytes", Message: "文件超过大小限制"}}})
		return
	}
	payload, err := io.ReadAll(io.LimitReader(file, h.maxUpload+1))
	if err != nil {
		writeError(c, err)
		return
	}
	render(c, stdhttp.StatusCreated, func() (domain.Bundle, error) {
		return h.bundles.Upload(c.Request.Context(), application.UploadCommand{Filename: header.Filename, Reader: bytes.NewReader(payload), IdempotencyKey: c.GetHeader("Idempotency-Key"), Actor: middleware.GetActor(c)})
	})
}

func (h *Handler) validateBundle(c *gin.Context) {
	render(c, stdhttp.StatusOK, func() (domain.Bundle, error) {
		return h.bundles.Validate(c.Request.Context(), c.Param("id"), middleware.GetActor(c))
	})
}

func (h *Handler) previewBundle(c *gin.Context) {
	render(c, stdhttp.StatusOK, func() (service.Preview, error) {
		return h.bundles.Preview(c.Request.Context(), c.Param("id"), middleware.GetActor(c))
	})
}

func (h *Handler) approveBundle(c *gin.Context) {
	render(c, stdhttp.StatusOK, func() (domain.Bundle, error) {
		return h.bundles.Approve(c.Request.Context(), c.Param("id"), middleware.GetActor(c))
	})
}

func (h *Handler) publishBundle(c *gin.Context) {
	request, ok := decode[struct {
		SelectedIDs []string `json:"selected_ids"`
	}](c)
	if !ok {
		return
	}
	render(c, stdhttp.StatusOK, func() (domain.Bundle, error) {
		return h.bundles.Publish(c.Request.Context(), application.PublishCommand{BundleID: c.Param("id"), SelectedIDs: request.SelectedIDs, IdempotencyKey: c.GetHeader("Idempotency-Key"), Actor: middleware.GetActor(c), RequestID: middleware.GetRequestID(c)})
	})
}

func (h *Handler) saveType(c *gin.Context) {
	item, ok := decode[domain.ConfigType](c)
	if !ok {
		return
	}
	if err := h.types.Save(c.Request.Context(), middleware.GetActor(c), item); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, item)
}

func (h *Handler) listTypes(c *gin.Context) {
	render(c, stdhttp.StatusOK, func() (gin.H, error) {
		items, err := h.types.List(c.Request.Context())
		return gin.H{"items": items, "total": len(items)}, err
	})
}

func (h *Handler) transition(c *gin.Context) {
	request, ok := decode[struct {
		Status   domain.ConfigStatus `json:"status"`
		Revision int64               `json:"revision"`
	}](c)
	if !ok {
		return
	}
	render(c, stdhttp.StatusOK, func() (domain.Configuration, error) {
		return h.lifecycle.Transition(c.Request.Context(), c.Param("id"), request.Status, request.Revision, middleware.GetActor(c), middleware.GetRequestID(c))
	})
}

func (h *Handler) rollback(c *gin.Context) {
	request, ok := decode[domain.RollbackRequest](c)
	if !ok {
		return
	}
	request.ConfigurationID = c.Param("id")
	render(c, stdhttp.StatusOK, func() (domain.Configuration, error) {
		return h.lifecycle.Rollback(c.Request.Context(), request, middleware.GetActor(c), middleware.GetRequestID(c))
	})
}

func (h *Handler) deleteConfiguration(c *gin.Context) {
	revision, err := strconv.ParseInt(c.Query("revision"), 10, 64)
	if err != nil {
		writeError(c, &domain.ValidationError{Code: "REVISION_REQUIRED", Fields: []domain.FieldError{{Path: "revision", Code: "integer", Message: "revision 必须是整数"}}})
		return
	}
	if err := h.lifecycle.Delete(c.Request.Context(), c.Param("id"), revision, middleware.GetActor(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(stdhttp.StatusNoContent)
}

func (h *Handler) impactReport(c *gin.Context) {
	render(c, stdhttp.StatusOK, func() (domain.ImpactReport, error) {
		return h.impact.Analyze(c.Request.Context(), c.Param("id"), middleware.GetActor(c))
	})
}

func (h *Handler) listAudits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	render(c, stdhttp.StatusOK, func() (application.AuditPage, error) {
		return h.audits.Browse(c.Request.Context(), middleware.GetActor(c), application.AuditQuery{
			Page: page, Size: size, Operation: c.Query("operation"), ActorID: c.Query("actor_id"), Target: c.Query("target"),
		})
	})
}

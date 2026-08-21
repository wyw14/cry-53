package health

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
)

type EndpointAdapter struct{ timeout time.Duration }

func NewEndpointAdapter(timeout time.Duration) *EndpointAdapter {
	return &EndpointAdapter{timeout: timeout}
}
func (a *EndpointAdapter) Name() string { return "local_endpoint" }

func (a *EndpointAdapter) Check(ctx context.Context, config domain.Configuration) domain.HealthResult {
	started := time.Now()
	result := domain.HealthResult{ConfigurationID: config.ID, Adapter: a.Name(), Status: domain.HealthHealthy, CheckedAt: started.UTC()}
	raw, ok := config.Values["endpoint"].(string)
	if !ok || raw == "" {
		result.Status, result.Message = domain.HealthDegraded, "配置未声明 endpoint"
		return result
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		result.Status, result.Message = domain.HealthUnhealthy, "endpoint 格式无效"
		return result
	}
	host := parsed.Host
	if parsed.Port() == "" {
		if parsed.Scheme == "https" {
			host = net.JoinHostPort(parsed.Hostname(), "443")
		} else {
			host = net.JoinHostPort(parsed.Hostname(), "80")
		}
	}
	timed, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	connection, err := (&net.Dialer{}).DialContext(timed, "tcp", host)
	result.Latency = time.Since(started)
	if err != nil {
		result.Status, result.Message = domain.HealthUnhealthy, fmt.Sprintf("本地连接失败: %v", err)
		return result
	}
	_ = connection.Close()
	result.Message = "本地连接成功"
	return result
}

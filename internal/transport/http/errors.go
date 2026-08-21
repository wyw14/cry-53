package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/middleware"
)

type errorResponse struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Fields    []domain.FieldError `json:"fields,omitempty"`
	RequestID string              `json:"request_id"`
}

func writeError(c *gin.Context, err error) {
	status, code, message := stdhttp.StatusInternalServerError, "INTERNAL", "服务发生内部错误"
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		status, code, message = stdhttp.StatusUnprocessableEntity, validation.Code, "请求内容未通过校验"
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = stdhttp.StatusNotFound, "NOT_FOUND", "资源不存在"
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = stdhttp.StatusForbidden, "FORBIDDEN", "当前身份无权执行该操作"
	case errors.Is(err, domain.ErrVersionConflict):
		status, code, message = stdhttp.StatusConflict, "VERSION_CONFLICT", "资源已被其他操作更新"
	case errors.Is(err, domain.ErrReferenced):
		status, code, message = stdhttp.StatusConflict, "CONFIG_REFERENCED", "配置仍被其他配置引用"
	case errors.Is(err, domain.ErrIdempotencyReuse):
		status, code, message = stdhttp.StatusConflict, "IDEMPOTENCY_REUSED", "幂等键已用于不同请求"
	case errors.Is(err, domain.ErrConflict):
		status, code, message = stdhttp.StatusConflict, "CONFLICT", "资源与现有数据冲突"
	case errors.Is(err, domain.ErrInvalidTransition):
		status, code, message = stdhttp.StatusConflict, "INVALID_TRANSITION", "当前状态不允许该操作"
	}
	response := errorResponse{Code: code, Message: message, RequestID: middleware.GetRequestID(c)}
	if validation != nil {
		response.Fields = validation.Fields
	}
	c.AbortWithStatusJSON(status, response)
}

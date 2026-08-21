package application

import (
	"context"
	"io"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New(string) string }
type Cipher interface {
	Encrypt(context.Context, string) (string, error)
	Decrypt(context.Context, string) (string, error)
}
type AttachmentStore interface {
	Save(context.Context, string, io.Reader) (string, error)
	Open(context.Context, string) (io.ReadCloser, error)
}
type Notifier interface {
	Notify(context.Context, string, map[string]string) error
}
type HealthAdapter interface {
	Name() string
	Check(context.Context, domain.Configuration) domain.HealthResult
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

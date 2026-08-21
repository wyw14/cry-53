package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	MasterKey       string
	AttachmentDir   string
	MaxUploadBytes  int64
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type Source interface {
	Lookup(string) (string, bool)
}

type EnvironmentSource struct{}

func (EnvironmentSource) Lookup(name string) (string, bool) {
	return os.LookupEnv(name)
}

func Load() (Config, error) {
	return LoadFrom(EnvironmentSource{})
}

func LoadFrom(source Source) (Config, error) {
	reader := decoder{source: source}
	settings := Config{
		HTTPAddr:        reader.text("HTTP_ADDR", ":8080"),
		DatabaseURL:     reader.text("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/config_registry?sslmode=disable"),
		MasterKey:       reader.text("MASTER_KEY", "0123456789abcdef0123456789abcdef"),
		AttachmentDir:   reader.text("ATTACHMENT_DIR", "./data/attachments"),
		MaxUploadBytes:  reader.bytes("MAX_UPLOAD_BYTES", 4<<20, 1<<10, 64<<20),
		RequestTimeout:  reader.duration("REQUEST_TIMEOUT", 5*time.Second),
		ShutdownTimeout: reader.duration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	reader.require(len(settings.MasterKey) == 32, "MASTER_KEY must contain exactly 32 bytes")
	reader.require(settings.RequestTimeout < settings.ShutdownTimeout, "REQUEST_TIMEOUT must be shorter than SHUTDOWN_TIMEOUT")
	return settings, errors.Join(reader.problems...)
}

type decoder struct {
	source   Source
	problems []error
}

func (d *decoder) text(name, fallback string) string {
	value, present := d.source.Lookup(name)
	if !present {
		return fallback
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func (d *decoder) bytes(name string, fallback, minimum, maximum int64) int64 {
	raw := d.text(name, strconv.FormatInt(fallback, 10))
	value, parseErr := strconv.ParseInt(raw, 10, 64)
	valid := parseErr == nil && value >= minimum && value <= maximum
	d.require(valid, fmt.Sprintf("%s must be between %d and %d", name, minimum, maximum))
	if !valid {
		return fallback
	}
	return value
}

func (d *decoder) duration(name string, fallback time.Duration) time.Duration {
	raw := d.text(name, fallback.String())
	value, parseErr := time.ParseDuration(raw)
	valid := parseErr == nil && value > 0
	d.require(valid, fmt.Sprintf("%s must be a positive duration", name))
	if !valid {
		return fallback
	}
	return value
}

func (d *decoder) require(condition bool, message string) {
	if !condition {
		d.problems = append(d.problems, errors.New(message))
	}
}

package domain

import "time"

type Usage struct {
	ID              string    `json:"id"`
	ConfigurationID string    `json:"configuration_id"`
	Consumer        string    `json:"consumer"`
	Environment     string    `json:"environment"`
	Criticality     string    `json:"criticality"`
	LastObservedAt  time.Time `json:"last_observed_at"`
}

type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnhealthy HealthStatus = "unhealthy"
)

type HealthResult struct {
	ConfigurationID string        `json:"configuration_id"`
	Adapter         string        `json:"adapter"`
	Status          HealthStatus  `json:"status"`
	Latency         time.Duration `json:"latency"`
	Message         string        `json:"message"`
	CheckedAt       time.Time     `json:"checked_at"`
}

type ImpactReport struct {
	ConfigurationID string         `json:"configuration_id"`
	Usages          []Usage        `json:"usages"`
	Health          []HealthResult `json:"health"`
	Blocking        bool           `json:"blocking"`
	Summary         string         `json:"summary"`
}

package wellknown

import "github.com/metagram-net/firehose/api"

type Handler struct{}

type HealthCheckResponse struct {
	Status string `json:"status"`
}

//nolint:unparam // The always-nil error is intentional.
func (Handler) HealthCheck(_ api.Context) (HealthCheckResponse, error) {
	return HealthCheckResponse{"OK"}, nil
}

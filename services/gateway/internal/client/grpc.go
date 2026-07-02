package client

import "fmt"

// Config holds gRPC connection settings for downstream services.
type Config struct {
	Addr string
}

func dialError(service string, err error) error {
	return fmt.Errorf("%s service request: %w", service, err)
}

func statusError(service string, code int, msg string) error {
	if msg != "" {
		return fmt.Errorf("%s service status %d: %s", service, code, msg)
	}
	return fmt.Errorf("%s service status %d", service, code)
}
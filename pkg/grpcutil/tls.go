package grpcutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// TLSConfig holds file paths for mutual TLS. Load via LoadTLSConfigFromEnv.
type TLSConfig struct {
	Enabled bool

	CAFile string

	ServerCertFile string
	ServerKeyFile  string

	ClientCertFile string
	ClientKeyFile  string

	// ServerName overrides TLS server name verification on client dials (SNI).
	ServerName string
}

// LoadTLSConfigFromEnv reads GRPC_MTLS_* / GRPC_TLS_* environment variables.
func LoadTLSConfigFromEnv() TLSConfig {
	v := os.Getenv("GRPC_MTLS_ENABLED")
	enabled := v == "true" || v == "1" || v == "yes"
	return TLSConfig{
		Enabled:        enabled,
		CAFile:         os.Getenv("GRPC_TLS_CA_FILE"),
		ServerCertFile: os.Getenv("GRPC_TLS_SERVER_CERT_FILE"),
		ServerKeyFile:  os.Getenv("GRPC_TLS_SERVER_KEY_FILE"),
		ClientCertFile: os.Getenv("GRPC_TLS_CLIENT_CERT_FILE"),
		ClientKeyFile:  os.Getenv("GRPC_TLS_CLIENT_KEY_FILE"),
		ServerName:     os.Getenv("GRPC_TLS_SERVER_NAME"),
	}
}

func (c TLSConfig) loadCAPool() (*x509.CertPool, error) {
	if c.CAFile == "" {
		return nil, fmt.Errorf("GRPC_TLS_CA_FILE is required when mTLS is enabled")
	}
	caPEM, err := os.ReadFile(c.CAFile)
	if err != nil {
		return nil, fmt.Errorf("read CA file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse CA file %s", c.CAFile)
	}
	return pool, nil
}

// ValidateServer checks server-side TLS configuration.
func (c TLSConfig) ValidateServer() error {
	if !c.Enabled {
		return nil
	}
	if c.ServerCertFile == "" || c.ServerKeyFile == "" {
		return fmt.Errorf("GRPC_MTLS_ENABLED requires GRPC_TLS_SERVER_CERT_FILE and GRPC_TLS_SERVER_KEY_FILE")
	}
	if c.CAFile == "" {
		return fmt.Errorf("GRPC_MTLS_ENABLED requires GRPC_TLS_CA_FILE")
	}
	if _, err := tls.LoadX509KeyPair(c.ServerCertFile, c.ServerKeyFile); err != nil {
		return fmt.Errorf("load server key pair: %w", err)
	}
	if _, err := c.loadCAPool(); err != nil {
		return err
	}
	return nil
}

// ValidateClient checks client-side TLS configuration.
func (c TLSConfig) ValidateClient() error {
	if !c.Enabled {
		return nil
	}
	if c.ClientCertFile == "" || c.ClientKeyFile == "" {
		return fmt.Errorf("GRPC_MTLS_ENABLED requires GRPC_TLS_CLIENT_CERT_FILE and GRPC_TLS_CLIENT_KEY_FILE")
	}
	if c.CAFile == "" {
		return fmt.Errorf("GRPC_MTLS_ENABLED requires GRPC_TLS_CA_FILE")
	}
	if _, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile); err != nil {
		return fmt.Errorf("load client key pair: %w", err)
	}
	if _, err := c.loadCAPool(); err != nil {
		return err
	}
	return nil
}

// ServerCredentials returns transport credentials for gRPC servers (requires client certs).
func (c TLSConfig) ServerCredentials() (credentials.TransportCredentials, error) {
	if err := c.ValidateServer(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(c.ServerCertFile, c.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server key pair: %w", err)
	}
	caPool, err := c.loadCAPool()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}), nil
}

// ClientCredentials returns transport credentials for gRPC clients (presents client cert).
func (c TLSConfig) ClientCredentials() (credentials.TransportCredentials, error) {
	if err := c.ValidateClient(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client key pair: %w", err)
	}
	caPool, err := c.loadCAPool()
	if err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}
	if c.ServerName != "" {
		tlsCfg.ServerName = c.ServerName
	}
	return credentials.NewTLS(tlsCfg), nil
}

// MTLSEnabled reports whether mutual TLS is active.
func MTLSEnabled() bool {
	return LoadTLSConfigFromEnv().Enabled
}
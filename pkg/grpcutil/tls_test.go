package grpcutil

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

func TestMTLSRoundTrip(t *testing.T) {
	dir := t.TempDir()
	material := generateMTLSMaterial(t)

	caFile := writeFile(t, dir, "ca.crt", material.caCert)
	serverCert := writeFile(t, dir, "server.crt", material.serverCert)
	serverKey := writeFile(t, dir, "server.key", material.serverKey)
	clientCert := writeFile(t, dir, "client.crt", material.clientCert)
	clientKey := writeFile(t, dir, "client.key", material.clientKey)

	t.Setenv("GRPC_MTLS_ENABLED", "true")
	t.Setenv("GRPC_TLS_CA_FILE", caFile)
	t.Setenv("GRPC_TLS_SERVER_CERT_FILE", serverCert)
	t.Setenv("GRPC_TLS_SERVER_KEY_FILE", serverKey)
	t.Setenv("GRPC_TLS_CLIENT_CERT_FILE", clientCert)
	t.Setenv("GRPC_TLS_CLIENT_KEY_FILE", clientKey)
	t.Setenv("GRPC_TLS_SERVER_NAME", "localhost")

	cfg := LoadTLSConfigFromEnv()
	serverCreds, err := cfg.ServerCredentials()
	if err != nil {
		t.Fatalf("server credentials: %v", err)
	}
	clientCreds, err := cfg.ClientCredentials()
	if err != nil {
		t.Fatalf("client credentials: %v", err)
	}

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	neobankv1.RegisterUserInternalServiceServer(grpcServer, &mtlsTestServer{})
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithChainUnaryInterceptor(correlationInterceptor),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer conn.Close()

	client := neobankv1.NewUserInternalServiceClient(conn)
	_, err = client.GetUserByID(context.Background(), &neobankv1.GetUserByIDRequest{UserId: "test"})
	if err != nil {
		t.Fatalf("rpc: %v", err)
	}
}

func TestMTLSRejectsClientWithoutCert(t *testing.T) {
	dir := t.TempDir()
	material := generateMTLSMaterial(t)

	caFile := writeFile(t, dir, "ca.crt", material.caCert)
	serverCert := writeFile(t, dir, "server.crt", material.serverCert)
	serverKey := writeFile(t, dir, "server.key", material.serverKey)

	cfg := TLSConfig{
		Enabled:        true,
		CAFile:         caFile,
		ServerCertFile: serverCert,
		ServerKeyFile:  serverKey,
	}
	serverCreds, err := cfg.ServerCredentials()
	if err != nil {
		t.Fatalf("server credentials: %v", err)
	}

	caPEM, _ := os.ReadFile(caFile)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	insecureClientCreds := credentials.NewTLS(&tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12})

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	neobankv1.RegisterUserInternalServiceServer(grpcServer, &mtlsTestServer{})
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecureClientCreds),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer conn.Close()

	client := neobankv1.NewUserInternalServiceClient(conn)
	_, err = client.GetUserByID(context.Background(), &neobankv1.GetUserByIDRequest{UserId: "test"})
	if err == nil {
		t.Fatal("expected TLS handshake failure without client cert")
	}
}

type mtlsTestServer struct {
	neobankv1.UnimplementedUserInternalServiceServer
}

func (s *mtlsTestServer) GetUserByID(context.Context, *neobankv1.GetUserByIDRequest) (*neobankv1.GetUserResponse, error) {
	return &neobankv1.GetUserResponse{
		User: &neobankv1.InternalUser{Id: "test", Status: "active"},
	}, nil
}

type mtlsMaterial struct {
	caCert, caKey       []byte
	serverCert, serverKey []byte
	clientCert, clientKey []byte
}

func generateMTLSMaterial(t *testing.T) mtlsMaterial {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("ca key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "neobank-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("ca cert: %v", err)
	}
	caCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyPEM := x509.MarshalPKCS1PrivateKey(caKey)

	serverCert, serverKey := signCert(t, caTmpl, caKey, "server", true, false)
	clientCert, clientKey := signCert(t, caTmpl, caKey, "client", false, true)

	return mtlsMaterial{
		caCert: caCert, caKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: caKeyPEM}),
		serverCert: serverCert, serverKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverKey}),
		clientCert: clientCert, clientKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: clientKey}),
	}
}

func signCert(t *testing.T, ca *x509.Certificate, caKey *rsa.PrivateKey, cn string, serverAuth, clientAuth bool) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	if serverAuth {
		tmpl.ExtKeyUsage = append(tmpl.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		tmpl.DNSNames = []string{"localhost"}
	}
	if clientAuth {
		tmpl.ExtKeyUsage = append(tmpl.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatalf("sign cert: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), x509.MarshalPKCS1PrivateKey(key)
}

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
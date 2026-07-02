#!/usr/bin/env bash
# Generate a dev/internal PKI for neobank gRPC mTLS.
# Output: ca.crt, server.{crt,key}, client.{crt,key}
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

DAYS=825
CA_SUBJ="${CA_SUBJ:-/CN=neobank-internal-ca}"
SERVER_SUBJ="${SERVER_SUBJ:-/CN=neobank-grpc-server}"
CLIENT_SUBJ="${CLIENT_SUBJ:-/CN=neobank-grpc-client}"

echo "Generating CA..."
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days "$DAYS" -out ca.crt -subj "$CA_SUBJ"

echo "Generating server key + CSR..."
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "$SERVER_SUBJ"

cat > server-ext.cnf <<'EOF'
subjectAltName = DNS:localhost,DNS:user,DNS:payment,DNS:card,DNS:notification,DNS:gateway,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days "$DAYS" -sha256 -extfile server-ext.cnf

echo "Generating client key + CSR..."
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "$CLIENT_SUBJ"

cat > client-ext.cnf <<'EOF'
extendedKeyUsage = clientAuth
EOF
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days "$DAYS" -sha256 -extfile client-ext.cnf

chmod 600 ca.key server.key client.key
rm -f server.csr client.csr server-ext.cnf client-ext.cnf ca.srl

cat <<EOF

Wrote:
  $DIR/ca.crt
  $DIR/server.crt / server.key
  $DIR/client.crt / client.key

Example env (all neobank gRPC servers + clients):

  export GRPC_MTLS_ENABLED=true
  export GRPC_TLS_CA_FILE=$DIR/ca.crt
  export GRPC_TLS_SERVER_CERT_FILE=$DIR/server.crt
  export GRPC_TLS_SERVER_KEY_FILE=$DIR/server.key
  export GRPC_TLS_CLIENT_CERT_FILE=$DIR/client.crt
  export GRPC_TLS_CLIENT_KEY_FILE=$DIR/client.key
  export GRPC_TLS_SERVER_NAME=localhost

goledger continues to use plaintext via grpcutil.DialInsecure.
EOF
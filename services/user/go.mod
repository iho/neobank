module github.com/iho/neobank/services/user

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/iho/neobank/pkg v0.0.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/oapi-codegen/runtime v1.4.2
	github.com/segmentio/kafka-go v0.4.51
	golang.org/x/crypto v0.46.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/redis/go-redis/v9 v9.21.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/iho/neobank/pkg => ../../pkg

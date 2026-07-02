module github.com/iho/neobank/services/card

go 1.24.0

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/iho/neobank/pkg v0.0.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/oapi-codegen/runtime v1.4.2
	google.golang.org/grpc v1.71.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/iho/neobank/pkg => ../../pkg

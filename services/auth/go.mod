module github.com/flick/backend/services/auth

go 1.24

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	go.uber.org/zap v1.27.0
	golang.org/x/oauth2 v0.30.0
	google.golang.org/grpc v1.71.1
	gorm.io/gorm v1.25.12
)

require (
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/flick/backend/pkg => ../../pkg

replace github.com/flick/backend/services/auth/proto => ./proto

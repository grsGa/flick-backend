module github.com/flick/backend/services/notification

go 1.24

require (
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.71.1
	gorm.io/gorm v1.25.12
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/flick/backend/pkg => ../../pkg

replace github.com/flick/backend/services/notification/proto => ./proto

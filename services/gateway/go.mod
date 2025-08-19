module github.com/flick/backend/services/gateway

go 1.24

require (
	github.com/99designs/gqlgen v0.17.78
	github.com/flick/backend/pkg/auth v0.0.0-00010101000000-000000000000
	github.com/flick/backend/pkg/config v0.0.0
	github.com/flick/backend/pkg/discovery v0.0.0-00010101000000-000000000000
	github.com/flick/backend/pkg/logger v0.0.0-00010101000000-000000000000
	github.com/flick/backend/pkg/telemetry v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/auth/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/content/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/interaction/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/media/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/messages/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/notification/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/recommendation/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/search/proto v0.0.0-00010101000000-000000000000
	github.com/flick/backend/services/user/proto v0.0.0-00010101000000-000000000000
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-gonic/gin v1.10.1
	github.com/hashicorp/consul/api v1.29.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/vektah/gqlparser/v2 v2.5.30
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.71.1
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/bytedance/sonic v1.13.3 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/arch v0.18.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250106144421-5f5ef82da422 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/flick/backend/pkg/auth => ../../pkg/auth

replace github.com/flick/backend/pkg/config => ../../pkg/config

replace github.com/flick/backend/pkg/database => ../../pkg/database

replace github.com/flick/backend/pkg/discovery => ../../pkg/discovery

replace github.com/flick/backend/pkg/logger => ../../pkg/logger

replace github.com/flick/backend/pkg/telemetry => ../../pkg/telemetry

replace github.com/flick/backend/services/auth/proto => ../auth/proto

replace github.com/flick/backend/services/content/proto => ../content/proto

replace github.com/flick/backend/services/interaction/proto => ../interaction/proto

replace github.com/flick/backend/services/media/proto => ../media/proto

replace github.com/flick/backend/services/messages/proto => ../messages/proto

replace github.com/flick/backend/services/notification/proto => ../notification/proto

replace github.com/flick/backend/services/recommendation/proto => ../recommendation/proto

replace github.com/flick/backend/services/search/proto => ../search/proto

replace github.com/flick/backend/services/user/proto => ../user/proto

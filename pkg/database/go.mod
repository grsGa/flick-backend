module github.com/flick/backend/pkg/database

go 1.21

require (
	github.com/flick/backend/pkg/config v0.0.0
	github.com/flick/backend/pkg/models v0.0.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

replace github.com/flick/backend/pkg/config => ../config
replace github.com/flick/backend/pkg/models => ../models
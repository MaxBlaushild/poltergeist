module github.com/MaxBlaushild/poltergeist/pkg/locationseeder

go 1.18

replace (
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../aws
	github.com/MaxBlaushild/poltergeist/pkg/db => ../db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/models => ../models
	github.com/MaxBlaushild/poltergeist/pkg/util => ../util
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.25.4 // indirect
)

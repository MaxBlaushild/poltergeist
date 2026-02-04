module github.com/MaxBlaushild/poltergeist/pkg/models

go 1.24.0

toolchain go1.24.10

require gorm.io/gorm v1.30.0 // direct

require (
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.11.1
	github.com/paulmach/orb v0.12.0
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96
	gorm.io/datatypes v1.2.7
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
)

replace (
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../aws
	github.com/MaxBlaushild/poltergeist/pkg/db => ../db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder => ../locationseeder
	github.com/MaxBlaushild/poltergeist/pkg/models => ../models
	github.com/MaxBlaushild/poltergeist/pkg/util => ../util
)

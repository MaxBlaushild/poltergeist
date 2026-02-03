module github.com/MaxBlaushild/poltergeist/pkg/db

go 1.24.0

toolchain go1.24.10

replace (
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../aws
	github.com/MaxBlaushild/poltergeist/pkg/db => ../db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder => ../locationseeder
	github.com/MaxBlaushild/poltergeist/pkg/models => ../models
	github.com/MaxBlaushild/poltergeist/pkg/util => ../util
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	gorm.io/driver/postgres v1.5.2
	gorm.io/gorm v1.30.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.11.1 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
)

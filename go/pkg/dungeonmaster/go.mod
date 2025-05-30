module github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster

go 1.22

toolchain go1.22.2

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
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/aws v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.38.5 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	googlemaps.github.io/maps v1.7.0 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.25.4 // indirect
)

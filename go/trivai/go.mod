module github.com/MaxBlaushild/poltergeist/trivai

go 1.24.0

toolchain go1.24.10

replace github.com/MaxBlaushild/poltergeist/pkg/texter => ../pkg/texter

replace github.com/MaxBlaushild/poltergeist/pkg/http => ../pkg/http

replace github.com/MaxBlaushild/poltergeist/final-fete => ../final-fete

replace github.com/MaxBlaushild/poltergeist/sonar => ../sonar

replace github.com/MaxBlaushild/poltergeist/travel-angels => ../travel-angels

replace github.com/MaxBlaushild/poltergeist/verifiable-sn => ../verifiable-sn

replace github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth

replace github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db

replace github.com/MaxBlaushild/poltergeist/pkg/hue => ../pkg/hue

replace github.com/MaxBlaushild/poltergeist/pkg/middleware => ../pkg/middleware

replace github.com/MaxBlaushild/poltergeist/pkg/models => ../pkg/models

replace github.com/MaxBlaushild/poltergeist/pkg/util => ../pkg/util

replace github.com/MaxBlaushild/poltergeist/pkg/liveness => ../pkg/liveness

replace github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../pkg/googlemaps

replace github.com/MaxBlaushild/poltergeist/pkg/aws => ../pkg/aws

replace github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest

replace github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster => ../pkg/dungeonmaster

replace github.com/MaxBlaushild/poltergeist/pkg/jobs => ../pkg/jobs

replace github.com/MaxBlaushild/poltergeist/pkg/locationseeder => ../pkg/locationseeder

replace github.com/MaxBlaushild/poltergeist/pkg/mapbox => ../pkg/mapbox

replace github.com/MaxBlaushild/poltergeist/pkg/useapi => ../pkg/useapi

replace github.com/MaxBlaushild/poltergeist/pkg/billing => ../pkg/billing

replace github.com/MaxBlaushild/poltergeist/pkg/dropbox => ../pkg/dropbox

replace github.com/MaxBlaushild/poltergeist/pkg/googledrive => ../pkg/googledrive

replace github.com/MaxBlaushild/poltergeist/pkg/cert => ../pkg/cert

replace github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor => ../pkg/ethereum_transactor

replace github.com/MaxBlaushild/poltergeist/pkg/twilio => ../pkg/twilio

replace github.com/MaxBlaushild/poltergeist/pkg/email => ../pkg/email

replace github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding

require (
	github.com/MaxBlaushild/poltergeist/pkg/auth v0.0.0-20230915032001-394ea26a68dd
	github.com/MaxBlaushild/poltergeist/pkg/billing v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/email v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/texter v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.6.0
	github.com/spf13/viper v1.16.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/http v0.0.0-00010101000000-000000000000 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lib/pq v1.11.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.13.0+incompatible // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.30.0 // indirect
)

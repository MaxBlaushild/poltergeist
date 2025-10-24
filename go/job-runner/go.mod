module github.com/MaxBlaushild/job-runner

go 1.22

replace (
	github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../pkg/aws
	github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster => ../pkg/dungeonmaster
	github.com/MaxBlaushild/poltergeist/pkg/email => ../pkg/email
	github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../pkg/googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/http => ../pkg/http
	github.com/MaxBlaushild/poltergeist/pkg/imagine => ../pkg/imagine
	github.com/MaxBlaushild/poltergeist/pkg/jobs => ../pkg/jobs
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder => ../pkg/locationseeder
	github.com/MaxBlaushild/poltergeist/pkg/models => ../pkg/models
	github.com/MaxBlaushild/poltergeist/pkg/slack => ../pkg/slack
	github.com/MaxBlaushild/poltergeist/pkg/texter => ../pkg/texter
	github.com/MaxBlaushild/poltergeist/pkg/twilio => ../pkg/twilio
	github.com/MaxBlaushild/poltergeist/pkg/useapi => ../pkg/useapi
	github.com/MaxBlaushild/poltergeist/pkg/util => ../pkg/util
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/aws v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0
	github.com/MaxBlaushild/poltergeist/pkg/jobs v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000
	github.com/disintegration/imaging v1.6.2
	github.com/google/uuid v1.6.0
	github.com/hibiken/asynq v0.25.0
	github.com/soniakeys/quant v1.0.0
	github.com/spf13/viper v1.19.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.30.0 // indirect
)

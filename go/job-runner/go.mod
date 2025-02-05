module github.com/MaxBlaushild/job-runner

go 1.22

replace (
	github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth
	github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/email => ../pkg/email
	github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding
	github.com/MaxBlaushild/poltergeist/pkg/http => ../pkg/http
	github.com/MaxBlaushild/poltergeist/pkg/imagine => ../pkg/imagine
	github.com/MaxBlaushild/poltergeist/pkg/jobs => ../pkg/jobs
	github.com/MaxBlaushild/poltergeist/pkg/models => ../pkg/models
	github.com/MaxBlaushild/poltergeist/pkg/slack => ../pkg/slack
	github.com/MaxBlaushild/poltergeist/pkg/texter => ../pkg/texter
	github.com/MaxBlaushild/poltergeist/pkg/twilio => ../pkg/twilio
	github.com/MaxBlaushild/poltergeist/pkg/useapi => ../pkg/useapi
	github.com/MaxBlaushild/poltergeist/pkg/util => ../pkg/util
)

require (
	cosmossdk.io/errors v1.0.1
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/imagine v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/jobs v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/useapi v0.0.0-00010101000000-000000000000
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/google/uuid v1.6.0
	github.com/hibiken/asynq v0.25.0
	github.com/spf13/viper v1.19.0
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
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
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240314234333-6e1732d8331c // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.25.4 // indirect
)

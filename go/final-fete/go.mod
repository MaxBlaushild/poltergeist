module github.com/MaxBlaushild/poltergeist/final-fete

go 1.24.0

replace (
	github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../pkg/aws
	github.com/MaxBlaushild/poltergeist/pkg/billing => ../pkg/billing
	github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/dropbox => ../pkg/dropbox
	github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster => ../pkg/dungeonmaster
	github.com/MaxBlaushild/poltergeist/pkg/email => ../pkg/email
	github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding
	github.com/MaxBlaushild/poltergeist/pkg/googledrive => ../pkg/googledrive
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../pkg/googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/http => ../pkg/http
	github.com/MaxBlaushild/poltergeist/pkg/hue => ../pkg/hue
	github.com/MaxBlaushild/poltergeist/pkg/imagine => ../pkg/imagine
	github.com/MaxBlaushild/poltergeist/pkg/jobs => ../pkg/jobs
	github.com/MaxBlaushild/poltergeist/pkg/liveness => ../pkg/liveness
	github.com/MaxBlaushild/poltergeist/pkg/location => ../pkg/location
	github.com/MaxBlaushild/poltergeist/pkg/locationseeder => ../pkg/locationseeder
	github.com/MaxBlaushild/poltergeist/pkg/mapbox => ../pkg/mapbox
	github.com/MaxBlaushild/poltergeist/pkg/middleware => ../pkg/middleware
	github.com/MaxBlaushild/poltergeist/pkg/models => ../pkg/models
	github.com/MaxBlaushild/poltergeist/pkg/slack => ../pkg/slack
	github.com/MaxBlaushild/poltergeist/pkg/texter => ../pkg/texter
	github.com/MaxBlaushild/poltergeist/pkg/twilio => ../pkg/twilio
	github.com/MaxBlaushild/poltergeist/pkg/useapi => ../pkg/useapi
	github.com/MaxBlaushild/poltergeist/pkg/util => ../pkg/util
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/auth v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/http v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/hue v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/liveness v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/middleware v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000 // indirect
	github.com/amimof/huego v1.2.1 // indirect
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/gin-gonic/gin v1.11.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.54.0 // indirect
	github.com/redis/go-redis/v9 v9.14.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	go.uber.org/mock v0.5.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/exp v0.0.0-20251125195548-87e1e737ad39 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.30.0 // indirect
)

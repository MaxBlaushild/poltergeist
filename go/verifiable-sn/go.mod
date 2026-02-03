module github.com/MaxBlaushild/poltergeist/verifiable-sn

go 1.24.0

toolchain go1.24.10

replace (
	github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../pkg/aws
	github.com/MaxBlaushild/poltergeist/pkg/billing => ../pkg/billing
	github.com/MaxBlaushild/poltergeist/pkg/cert => ../pkg/cert
	github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/dropbox => ../pkg/dropbox
	github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster => ../pkg/dungeonmaster
	github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding
	github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor => ../pkg/ethereum_transactor
	github.com/MaxBlaushild/poltergeist/pkg/googledrive => ../pkg/googledrive
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps => ../pkg/googlemaps
	github.com/MaxBlaushild/poltergeist/pkg/http => ../pkg/http
	github.com/MaxBlaushild/poltergeist/pkg/jobs => ../pkg/jobs
	github.com/MaxBlaushild/poltergeist/pkg/liveness => ../pkg/liveness
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
	firebase.google.com/go/v4 v4.14.0
	github.com/MaxBlaushild/poltergeist/pkg/auth v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/aws v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/cert v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/db v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/middleware v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/util v0.0.0-00010101000000-000000000000
	github.com/ethereum/go-ethereum v1.14.0
	github.com/fxamacker/cbor/v2 v2.5.0
	github.com/gin-gonic/gin v1.11.0
	github.com/google/uuid v1.6.0
	github.com/spf13/viper v1.21.0
	google.golang.org/api v0.170.0
)

require (
	cloud.google.com/go v0.112.1 // indirect
	cloud.google.com/go/compute v1.24.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/firestore v1.15.0 // indirect
	cloud.google.com/go/iam v1.1.7 // indirect
	cloud.google.com/go/longrunning v0.5.5 // indirect
	cloud.google.com/go/storage v1.40.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/googlemaps v0.0.0 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/http v0.0.0-00010101000000-000000000000 // indirect
	github.com/MaxBlaushild/poltergeist/pkg/liveness v0.0.0-00010101000000-000000000000 // indirect
	github.com/MicahParks/keyfunc v1.9.0 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.11.1 // indirect
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
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/mock v0.5.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/appengine/v2 v2.0.2 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240314234333-6e1732d8331c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240311132316-a219d84964c2 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/driver/postgres v1.5.2 // indirect
	gorm.io/gorm v1.30.0 // indirect
)

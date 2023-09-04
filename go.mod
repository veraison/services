module github.com/veraison/services

go 1.19

replace github.com/veraison/services => github.com/Kakemone/services v0.0.0-20230810131214-bda013b5a960

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/fatih/color v1.15.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-playground/assert/v2 v2.2.0
	github.com/golang/mock v1.6.0
	github.com/google/go-tpm v0.9.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-plugin v1.4.10
	github.com/jellydator/ttlcache/v3 v3.0.1
	github.com/lestrrat-go/jwx/v2 v2.0.11
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/mitchellh/mapstructure v1.5.0
	github.com/moogar0880/problems v0.1.1
	github.com/open-policy-agent/opa v0.55.0
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20230725210150-fb29fc3c913e
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.7.0
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
	github.com/veraison/ccatoken v1.1.0
	github.com/veraison/cmw v0.1.0
	github.com/veraison/corim v1.1.2-0.20230904093201-6df0917f7caf
	github.com/veraison/dice v0.0.1
	github.com/veraison/ear v1.1.0
	github.com/veraison/eat v0.0.0-20220117140849-ddaf59d69f53
	github.com/veraison/go-cose v1.1.1-0.20230613195103-433d4c233485
	github.com/veraison/parsec v0.1.0
	github.com/veraison/psatoken v1.2.0
	go.uber.org/zap v1.25.0
	golang.org/x/text v0.12.0
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.10.0-rc3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.15.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.4 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230726155614-23370e0ffb3e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230807174057-1744710a1577 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/spf13/pflag v1.0.5
	github.com/veraison/swid v1.1.0
	go.uber.org/multierr v1.11.0 // indirect
)

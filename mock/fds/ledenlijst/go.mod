module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/fds/ledenlijst

go 1.23.0

toolchain go1.23.2

require (
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers v0.0.0-20250205102328-4cadc1a4635f
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server v0.0.0-20250401145057-c176bf65f90e
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-20250401145057-c176bf65f90e
	gitlab.com/gjuyn/go-config v1.2.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/defensestation/osquery v1.0.0 // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-co-op/gocron/v2 v2.15.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/kvtools/etcdv3 v1.0.2 // indirect
	github.com/kvtools/valkeyrie v1.0.0 // indirect
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/opensearch-project/opensearch-go v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.58.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components v0.0.0-20250401145057-c176bf65f90e // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models v0.0.0-20250401145057-c176bf65f90e // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas v0.0.0-20250401145057-c176bf65f90e // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci v0.0.0-20250401145057-c176bf65f90e // indirect
	go.etcd.io/etcd/api/v3 v3.5.12 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.12 // indirect
	go.etcd.io/etcd/client/v3 v3.5.12 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20250210185358-939b2ce775ac // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.32.2 // indirect
	k8s.io/client-go v0.32.2 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components => ./../../../eam/components
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers => ./../../../eam/handlers
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models => ./../../../eam/models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server => ./../../../eam/server
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas => ../../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../../utilities
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci => ../../../utilities-no-ci
)

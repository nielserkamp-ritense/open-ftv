module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip

go 1.23.8

toolchain go1.24.2

require (
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-co-op/gocron/v2 v2.16.1
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.17.1
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/kvtools/etcdv3 v1.0.2
	github.com/kvtools/valkeyrie v1.0.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/mimetype v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-20250415141202-eea5915a6d13
	k8s.io/client-go v0.32.3
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.60.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas v0.0.0-20250410082626-e73c066519bd // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/v3 v3.5.4 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250414145226-207652e42e2e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250414145226-207652e42e2e // indirect
	google.golang.org/grpc v1.71.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.32.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/mimetype => ../mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server => ../server
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas => ../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../utilities
)

module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip

go 1.24.6

require (
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-co-op/gocron/v2 v2.16.2
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.18.0
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/kvtools/etcdv3 v1.0.2
	github.com/kvtools/valkeyrie v1.0.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20250723072950-f3dd075b1605
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250806094001-89fc982af2f3
	k8s.io/client-go v0.33.3
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
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
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.63.0 // indirect
	go.etcd.io/etcd/api/v3 v3.6.1 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.1 // indirect
	go.etcd.io/etcd/client/v3 v3.6.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.33.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype => ../mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server => ../server
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas => ../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../utilities
)

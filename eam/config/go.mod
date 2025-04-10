module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/config

go 1.23.0

toolchain go1.23.2

require (
	github.com/kvtools/consul v1.0.2
	github.com/kvtools/etcdv3 v1.0.2
	github.com/kvtools/valkeyrie v1.0.0
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-00010101000000-000000000000
	gitlab.com/gjuyn/go-config v1.2.0
)

require (
	github.com/armon/go-metrics v0.3.10 // indirect
	github.com/cedar-policy/cedar-go v1.1.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-co-op/gocron/v2 v2.15.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/consul/api v1.14.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.9.7 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.4 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas v0.0.0-00010101000000-000000000000 // indirect
	go.etcd.io/etcd/api/v3 v3.5.12 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.12 // indirect
	go.etcd.io/etcd/client/v3 v3.5.12 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.71.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.32.2 // indirect
	k8s.io/client-go v0.32.2 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components => ../components
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server => ../server
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas => ../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../utilities
)

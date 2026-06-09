module github.com/txbao/goeasy-cli

go 1.25.0

require (
	github.com/go-sql-driver/mysql v1.10.0
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/jackc/pgx/v5 v5.10.0
	github.com/spf13/cobra v1.8.1
	github.com/txbao/goeasy v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/txbao/goeasy => ../goeasy

exclude google.golang.org/genproto v0.0.0-20220314164441-57ef72a4c106

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/nsqio/go-nsq v1.1.0 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.etcd.io/etcd/api/v3 v3.6.12 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.12 // indirect
	go.etcd.io/etcd/client/v3 v3.6.12 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

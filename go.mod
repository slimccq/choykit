module devpkg.work/choykit

go 1.15

// etcd go module的坑
// https://colobu.com/2020/04/09/accidents-of-etcd-and-go-module/
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v1.8.4
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/jessevdk/go-flags v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/templexxx/xorsimd v0.4.1
	github.com/tjfoc/gmsm v1.4.0
	go.etcd.io/etcd v3.3.25+incompatible
	go.mongodb.org/mongo-driver v1.5.1
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc
	golang.org/x/sys v0.0.0-20210419170143-37df388d1f33
	google.golang.org/protobuf v1.26.0
)

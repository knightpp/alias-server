module github.com/knightpp/alias-server

go 1.20

require (
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/huandu/go-clone/generic v1.5.1
	github.com/knightpp/alias-proto/go v0.0.0-20230203213652-4df0cf3b0680
	github.com/life4/genesis v1.1.0
	github.com/onsi/ginkgo/v2 v2.8.3
	github.com/onsi/gomega v1.27.1
	github.com/redis/go-redis/v9 v9.0.2
	github.com/rs/zerolog v1.29.0
	github.com/stretchr/testify v1.8.2
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/huandu/go-clone v1.5.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230223222841-637eb2293923 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/knightpp/alias-proto/go => ../alias-proto/go

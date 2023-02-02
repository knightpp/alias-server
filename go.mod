module github.com/knightpp/alias-server

go 1.20

require (
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/huandu/go-clone/generic v1.4.0
	github.com/knightpp/alias-proto/go v0.0.0-20230131161028-d77d03b7020f
	github.com/onsi/gomega v1.26.0
	github.com/redis/go-redis/v9 v9.0.2
	github.com/rs/zerolog v1.29.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/sync v0.1.0
	google.golang.org/grpc v1.52.3
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/huandu/go-clone v1.4.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230131230820-1c016267d619 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/knightpp/alias-proto/go => ../alias-proto/go

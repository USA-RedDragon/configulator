module github.com/USA-RedDragon/configulator/v2/examples/all

go 1.27

require (
	github.com/USA-RedDragon/configulator/v2 v2.0.0
	github.com/goccy/go-yaml v1.19.2
	github.com/spf13/cobra v1.10.1
	github.com/spf13/pflag v1.0.10
)

require (
	github.com/dave/jennifer v1.7.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
)

replace github.com/USA-RedDragon/configulator/v2 => ../..

tool github.com/USA-RedDragon/configulator/v2/cmd/configulator

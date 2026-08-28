module github.com/USA-RedDragon/configulator/v2/examples/all

go 1.27

require (
	github.com/USA-RedDragon/configulator/v2 v2.0.0
	github.com/goccy/go-yaml v1.19.2
	github.com/spf13/cobra v1.10.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
)

replace github.com/USA-RedDragon/configulator/v2 => ../..

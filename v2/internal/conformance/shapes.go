//go:build goexperiment.jsonv2

// Package conformance holds the five SPEC shapes (spec/shapes.md) plus the
// extended Gate A cases (embedded struct, UnmarshalJSONFrom). The
// *_configulator.go files here are HAND-WRITTEN in exactly the form the
// generator must emit: they are the generator's golden fixture and the
// corpus runner's shape implementations at once, so the hand-written work
// is the test oracle, not throwaway.
package conformance

import "time"

// --- shape: scalars ---

type Scalars struct {
	Name    string  `name:"name" json:"name" default:"svc"`
	Count   int64   `name:"count" json:"count"`
	Port    uint16  `name:"port" json:"port" default:"8080"`
	Ratio   float64 `name:"ratio" json:"ratio" default:"1.5"`
	Verbose bool    `name:"verbose" json:"verbose" default:"false"`
}

func (Scalars) Validate() error { return nil }

// --- shape: nested ---

type Nested struct {
	AppName string `name:"app-name" json:"app-name" default:"myapp"`
	HTTP    NHTTP  `name:"http" json:"http"`
	DB      NDB    `name:"db" json:"db"`
}

type NHTTP struct {
	Host string `name:"host" json:"host" default:"localhost"`
	Port uint16 `name:"port" json:"port" default:"8080"`
}

type NDB struct {
	URL  string `name:"url" json:"url" default:"postgres://localhost/db"`
	Pool NPool  `name:"pool" json:"pool"`
}

type NPool struct {
	Size uint16 `name:"size" json:"size" default:"10"`
}

func (Nested) Validate() error { return nil }

// --- shape: collections ---

type Collections struct {
	Tags     []string          `name:"tags" json:"tags" default:"a,b"`
	Labels   map[string]string `name:"labels" json:"labels"`
	Servers  []Server          `name:"servers" json:"servers"`
	Pools    map[string]Pool   `name:"pools" json:"pools"`
	LogLevel string            `name:"log-level" json:"log-level" default:"info"`
}

type Server struct {
	Addr   string `name:"addr" json:"addr"`
	Weight uint16 `name:"weight" json:"weight" default:"1"`
}

type Pool struct {
	Size uint16 `name:"size" json:"size" default:"5"`
}

func (Collections) Validate() error { return nil }

// --- shape: optionals ---

type Optionals struct {
	Port *uint16 `name:"port" json:"port"`
	Name *string `name:"name" json:"name" default:"opt-name"`
	TLS  *Tls    `name:"tls" json:"tls"`
}

type Tls struct {
	Cert       string `name:"cert" json:"cert"`
	MinVersion uint16 `name:"min-version" json:"min-version" default:"12"`
}

func (Optionals) Validate() error { return nil }

// --- shape: durations ---

type Durations struct {
	Timeout time.Duration `name:"timeout" json:"timeout" default:"30s"`
	Label   string        `name:"label" json:"label"`
}

func (Durations) Validate() error { return nil }

// --- golden-only: embedded struct (flattened into the parent namespace) ---

type Base struct {
	Region string `name:"region" json:"region" default:"us-east-1"`
}

type Embedded struct {
	Base        // flattened: "region" is a top-level key
	Zone string `name:"zone" json:"zone" default:"a"`
}

func (Embedded) Validate() error { return nil }

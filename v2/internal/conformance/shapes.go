//go:build goexperiment.jsonv2

// Package conformance holds the spec shapes (spec/shapes.md) plus
// extra cases for embedded structs and UnmarshalJSONFrom. The
// *_configulator.go files here are both the generator's golden fixture and
// the shape implementations the corpus runner uses.
package conformance

import "time"

type Scalars struct {
	Name    string  `name:"name" json:"name" default:"svc"`
	Count   int64   `name:"count" json:"count"`
	Port    uint16  `name:"port" json:"port" default:"8080"`
	Ratio   float64 `name:"ratio" json:"ratio" default:"1.5"`
	Verbose bool    `name:"verbose" json:"verbose" default:"false"`
}

func (Scalars) Validate() error { return nil }

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

type NestedCollections struct {
	Peers  []NCPeer          `name:"peers" json:"peers"`
	ByName map[string]NCPeer `name:"by-name" json:"by-name"`
}

type NCPeer struct {
	Name  string            `name:"name" json:"name"`
	Slots int64             `name:"slots" json:"slots" default:"3"`
	Rules []NCRule          `name:"rules" json:"rules"`
	Tags  map[string]NCRule `name:"tags" json:"tags"`
	Inner NCLevel           `name:"inner" json:"inner"`
	Opt   *NCLevel          `name:"opt" json:"opt"`
}

type NCRule struct {
	From  int64 `name:"from" json:"from"`
	Range int64 `name:"range" json:"range" default:"1"`
	On    bool  `name:"on" json:"on" default:"true"`
}

type NCLevel struct {
	Level int64 `name:"level" json:"level" default:"7"`
}

func (NestedCollections) Validate() error { return nil }

type Optionals struct {
	Port *uint16    `name:"port" json:"port"`
	Name *string    `name:"name" json:"name" default:"opt-name"`
	TLS  *TLSConfig `name:"tls" json:"tls"`
}

type TLSConfig struct {
	Cert       string `name:"cert" json:"cert"`
	MinVersion uint16 `name:"min-version" json:"min-version" default:"12"`
}

func (Optionals) Validate() error { return nil }

type Durations struct {
	Timeout time.Duration `name:"timeout" json:"timeout" default:"30s"`
	Label   string        `name:"label" json:"label"`
}

func (Durations) Validate() error { return nil }

type Complex struct {
	Z        complex128   `name:"z" json:"z" default:"1+2i"`
	Exponent complex128   `name:"exponent" json:"exponent"`
	W        complex64    `name:"w" json:"w"`
	Zs       []complex128 `name:"zs" json:"zs"`
}

func (Complex) Validate() error { return nil }

// golden fixture only, not a corpus shape

type Base struct {
	Region string `name:"region" json:"region" default:"us-east-1"`
}

type Embedded struct {
	Base        // flattened: "region" is a top-level key
	Zone string `name:"zone" json:"zone" default:"a"`
}

func (Embedded) Validate() error { return nil }

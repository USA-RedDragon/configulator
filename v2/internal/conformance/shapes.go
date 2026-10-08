//go:build goexperiment.jsonv2

// Package conformance holds the spec shapes (spec/shapes.md) plus
// extra cases for embedded structs and UnmarshalJSONFrom. The
// *_configulator.go files here are both the generator's golden fixture and
// the shape implementations the corpus runner uses.
package conformance

import (
	"errors"
	"time"
)

// Scalars is the scalars shape: one field of each scalar kind.
type Scalars struct {
	Name    string  `name:"name" json:"name" default:"svc"`
	Count   int64   `name:"count" json:"count"`
	Port    uint16  `name:"port" json:"port" default:"8080"`
	Ratio   float64 `name:"ratio" json:"ratio" default:"1.5"`
	Verbose bool    `name:"verbose" json:"verbose" default:"false"`
}

// Validate implements configulator.Validator.
func (Scalars) Validate() error { return nil }

// Nested is the nested shape: two levels of nested structs.
type Nested struct {
	AppName string `name:"app-name" json:"app-name" default:"myapp"`
	HTTP    NHTTP  `name:"http" json:"http"`
	DB      NDB    `name:"db" json:"db"`
}

// NHTTP is the http section of Nested.
type NHTTP struct {
	Host string `name:"host" json:"host" default:"localhost"`
	Port uint16 `name:"port" json:"port" default:"8080"`
}

// NDB is the db section of Nested.
type NDB struct {
	URL  string `name:"url" json:"url" default:"postgres://localhost/db"`
	Pool NPool  `name:"pool" json:"pool"`
}

// NPool is the db.pool section of Nested.
type NPool struct {
	Size uint16 `name:"size" json:"size" default:"10"`
}

// Validate implements configulator.Validator.
func (Nested) Validate() error { return nil }

// Collections is the collections shape: lists and maps of scalars and structs.
type Collections struct {
	Tags     []string          `name:"tags" json:"tags" default:"a,b"`
	Labels   map[string]string `name:"labels" json:"labels"`
	Servers  []Server          `name:"servers" json:"servers"`
	Pools    map[string]Pool   `name:"pools" json:"pools"`
	LogLevel string            `name:"log-level" json:"log-level" default:"info"`
}

// Server is an element of Collections.Servers.
type Server struct {
	Addr   string `name:"addr" json:"addr"`
	Weight uint16 `name:"weight" json:"weight" default:"1"`
}

// Pool is an element of Collections.Pools.
type Pool struct {
	Size uint16 `name:"size" json:"size" default:"5"`
}

// Validate implements configulator.Validator.
func (Collections) Validate() error { return nil }

// NestedCollections is the nested-collections shape: collections inside collection elements.
type NestedCollections struct {
	Peers  []NCPeer          `name:"peers" json:"peers"`
	ByName map[string]NCPeer `name:"by-name" json:"by-name"`
}

// NCPeer is an element of NestedCollections.Peers and ByName.
type NCPeer struct {
	Name  string            `name:"name" json:"name"`
	Slots int64             `name:"slots" json:"slots" default:"3"`
	Rules []NCRule          `name:"rules" json:"rules"`
	Tags  map[string]NCRule `name:"tags" json:"tags"`
	Inner NCLevel           `name:"inner" json:"inner"`
	Opt   *NCLevel          `name:"opt" json:"opt"`
}

// NCRule is an element of NCPeer.Rules and Tags.
type NCRule struct {
	From  int64 `name:"from" json:"from"`
	Range int64 `name:"range" json:"range" default:"1"`
	On    bool  `name:"on" json:"on" default:"true"`
}

// NCLevel is the struct under NCPeer.Inner and Opt.
type NCLevel struct {
	Level int64 `name:"level" json:"level" default:"7"`
}

// Validate implements configulator.Validator.
func (NestedCollections) Validate() error { return nil }

// Optionals is the optionals shape: optional scalars and an optional struct.
type Optionals struct {
	Port *uint16    `name:"port" json:"port"`
	Name *string    `name:"name" json:"name" default:"opt-name"`
	TLS  *TLSConfig `name:"tls" json:"tls"`
}

// TLSConfig is the optional tls section of Optionals.
type TLSConfig struct {
	Cert       string `name:"cert" json:"cert"`
	MinVersion uint16 `name:"min-version" json:"min-version" default:"12"`
}

// Validate implements configulator.Validator.
func (Optionals) Validate() error { return nil }

// Durations is the durations shape.
type Durations struct {
	Timeout time.Duration `name:"timeout" json:"timeout" default:"30s"`
	Label   string        `name:"label" json:"label"`
}

// Validate implements configulator.Validator.
func (Durations) Validate() error { return nil }

// Complex is the complex shape: complex numbers and a list of them.
type Complex struct {
	Z        complex128   `name:"z" json:"z" default:"1+2i"`
	Exponent complex128   `name:"exponent" json:"exponent"`
	W        complex64    `name:"w" json:"w"`
	Zs       []complex128 `name:"zs" json:"zs"`
}

// Validate implements configulator.Validator.
func (Complex) Validate() error { return nil }

// Required is the required shape: required fields at every level, plus a validation hook.
type Required struct {
	Top    string  `name:"top" json:"top" required:"true"`
	Nested RNested `name:"nested" json:"nested"`
	Opt    *ROpt   `name:"opt" json:"opt"`
	Items  []ROpt  `name:"items" json:"items"`
}

// RNested is the nested section of Required.
type RNested struct {
	Leaf string `name:"leaf" json:"leaf" required:"true"`
}

// ROpt is the optional section and the list element of Required.
type ROpt struct {
	Leaf  string `name:"leaf" json:"leaf" required:"true"`
	Other string `name:"other" json:"other"`
}

// Validate implements configulator.Validator.
func (r Required) Validate() error {
	if r.Top == "invalid" {
		return errors.New("top must not be invalid")
	}
	return nil
}

// Attributes is the attributes shape: a secret, renames, skips and a shorthand.
type Attributes struct {
	Token   int64  `name:"token" json:"token" secret:"true"`
	Renamed string `name:"renamed" json:"renamed" env:"RN" flag:"rn"`
	NoEnv   string `name:"no-env" json:"no-env" env:"-"`
	NoFlag  string `name:"no-flag" json:"no-flag" flag:"-"`
	Port    uint16 `name:"port" json:"port" short:"p"`
}

// Validate implements configulator.Validator.
func (Attributes) Validate() error { return nil }

// Base is embedded in Embedded, so its fields are top-level keys.
type Base struct {
	Region string `name:"region" json:"region" default:"us-east-1"`
}

// Embedded tests an embedded struct. It is a golden fixture, not a corpus shape.
type Embedded struct {
	Base

	Zone string `name:"zone" json:"zone" default:"a"`
}

// Validate implements configulator.Validator.
func (Embedded) Validate() error { return nil }

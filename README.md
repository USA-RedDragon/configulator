# Configulator

[![go.mod version](https://img.shields.io/github/go-mod/go-version/USA-RedDragon/configulator.svg)](https://github.com/USA-RedDragon/configulator) [![codecov](https://codecov.io/gh/USA-RedDragon/configulator/graph/badge.svg?token=AhUJaQtw9R)](https://codecov.io/gh/USA-RedDragon/configulator) [![License](https://badgen.net/github/license/USA-RedDragon/configulator)](https://github.com/USA-RedDragon/configulator/blob/main/LICENSE) [![GitHub contributors](https://badgen.net/github/contributors/USA-RedDragon/configulator)](https://github.com/USA-RedDragon/configulator/graphs/contributors/) [![GoReportCard example](https://goreportcard.com/badge/github.com/USA-RedDragon/configulator)](https://goreportcard.com/report/github.com/USA-RedDragon/configulator)

A simple configuration manager for use in my apps

## Features

- Supports configuration from:
  - YAML or JSON files
  - Environment variables
  - Command line arguments (`spf13/pflag`)

## Supported types

Working:

- all scalars except complex
- structs
- arrays of scalars

Not working:

- maps
- arrays of maps
- arrays of structs
- multi-dimensional arrays
- complex scalars

## Usage

> [!NOTE]
Because the configuration options are expressed as different cases (i.e. `http.host` in YAML would be `HTTP__HOST` in environment variables), this library cannot be used for configurations that contain the same field name in different cases.

### Struct tags

This library uses the `name`, `default`, and `description` tags. Multiple values can be passed if they are comma separated. Fields without a `name` tag are not utilized. The field name can be inferred from `json` or `yaml` tags if present.

Examples of struct field tags and their meanings:

```go
// Field appears in config files, environment variables, and command line arguments as key "myName".
Field int `name:"myName"`

// Field has a description if seen in the CLI's --help
Field int `name:"myName" description:"this text appears in the help section of the CLI"`

// Field has a default value of 1
Field int `name:"myName" default:"1"`
```

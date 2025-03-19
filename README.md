# Configulator

A simple configuration manager for use in my apps

## Features

- Supports configuration from:
  - YAML or JSON files
  - Environment variables
  - Command line arguments (`spf13/pflag`)

## Usage

> [!NOTE]
Because the configuration options are expressed as different cases (i.e. `http.host` in YAML would be `HTTP__HOST` in environment variables), this library cannot be used for configurations that contain the same field name in different cases.

### Struct tags

This library uses the `config` tag. Multiple values can be passed if they are comma separated. Fields without a config tag are not counted

Examples of struct field tags and their meanings:

```go
// Field appears in config files, environment variables, and command line arguments as key "myName".
Field int `config:"myName"`

// Same as above, but without the shorthand
Field int `config:"name:myName"`

// Field has a description if seen in the CLI's --help
Field int `config:"myName,description:this text appears in the help section of the CLI"`

// Field has a default value of 1
Field int `config:"myName,default:1"`
```

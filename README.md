# ConfigReader

[![Build](https://github.com/go-srv/configreader/actions/workflows/workflow.yml/badge.svg)](https://github.com/go-srv/configreader/actions/workflows/workflow.yml)
[![codecov](https://codecov.io/gh/go-srv/configreader/branch/master/graph/badge.svg?token=EGDJ3BOH2Y)](https://codecov.io/gh/go-srv/configreader)

## Introduction

Configreader is a package wraps [spf13/viper](https://github.com/spf13/viper)
to read, merge config files, environments values, application flags, and default values defined in the tags of the struct.

## Usage

### Tags Of Struct

* **key** defines the name of the field in a config file
* **default** defines the default of the field, if there is no value provided in the file/env/flags, the default value will be used
* **flag** defines the flag name for the field
* **env** defines the environment variable name for the field, a default env previs `APP_` will be added
* **required** defines if the field is required, if the field is required, and there is no value provided, an error will occured
* **validation** defines simple methods to validate the value of the field.

TODO: explain the tag details here

## Usages

```go
import "github.com/go-srv/configreader"

type Config struct {
    Host string `key:"host" flag:"host" env:"host" required:"true" validation:"range:[8:255]"`
    Port int `key:"port" required:"true" flag:"port" env:"port" validation:"range:[80:65535]"`
    LogLevel string `key:"loglevel" default:"info" validation:"in:[error, warning, info, verbose]"`
}

func main() {
    c := Config{}
    configreader.LoadFromFile("/path/to/config/file.ext", &c)
}
```

More usages please refer to the test file.

TODO: add more usages here

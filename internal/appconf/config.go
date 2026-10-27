package appconf

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

const (
	appEnvFilePath = "."
	appEnvFileName = ".env"
	appEnvFileExt  = "env"
)

type validator interface {
	Validate() error
}

// Load reads config from .env file (if present) and environment variables into T and validates it.
// T must be a struct whose fields are config sections implementing Validate with a value receiver
func Load[T any]() (*T, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.AddConfigPath(appEnvFilePath)
	v.SetConfigName(appEnvFileName)
	v.SetConfigType(appEnvFileExt)

	if err := v.ReadInConfig(); err != nil {
		// if file not found use env variables only
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, err
		}
	}

	v.AutomaticEnv()

	var c T
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	if err := Validate(c); err != nil {
		return nil, err
	}

	return &c, nil
}

// Validate validates every section of config struct c. Every field of c must implement Validate with a value receiver
func Validate(c any) error {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct, got %s", v.Kind())
	}

	var errs []error
	for i := range v.NumField() {
		field := v.Field(i)
		// unexported fields can't be type asserted
		var section validator
		ok := field.CanInterface()
		if ok {
			section, ok = reflect.TypeAssert[validator](field)
		}
		if !ok {
			// a section without Validate is a programming error, not an optional section
			return fmt.Errorf("config section %s has no Validate method", v.Type().Field(i).Name)
		}
		errs = append(errs, section.Validate())
	}

	return errors.Join(errs...)
}

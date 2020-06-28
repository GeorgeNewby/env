// Package env implements a simple .env file parser to set environment variables.
// It is also used to unmarshal environment variables to structs.
package env

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Load sets environment variables defined in the .env file.
// Errors are only generated if the file exists.
func Load() error {
	file, err := os.Open(".env")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error opening env file: %w", err)
	}
	defer file.Close()

	return Parse(file)
}

// Parse sets environment variables defined in the reader.
// The format is 'FOO=bar' with '#' used for comments.
// The format is shown bellow:
//  # This is a comment
//  FOO=foo
//
//  # Leading and trailing spaces are ignored
//  BAR = bar
//
//  # Sentences are allowed
//  BAZ=foo bar baz
func Parse(r io.Reader) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		pair := strings.Split(line, "=")
		if len(pair) != 2 {
			return fmt.Errorf("invalid variable format: %s", pair)
		}
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error settings variable %s=%s: %w", key, value, err)
		}
	}
	return nil
}

// Unmarshal populates the fields of a struct pointed to by v with the corresponding environment variables.
// Only variables with the 'env' struct tag are populated:
//  struct {
//      Foo string  `env:"FOO"`
//      Bar int     `env:"BAR"`
//      Baz float64 `env:"BAZ"`
//      Qux bool    `env:"QUX"`
//  }
// The types supported are strings, ints, floats and bools.
// An error is thrown is the variable doesn't exist.
func Unmarshal(v interface{}) error {
	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Ptr {
		return fmt.Errorf("expected struct pointer, got %T", v)
	}

	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct pointer, got %T", v)
	}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		name, ok := s.Type().Field(i).Tag.Lookup("env")
		if ok && f.CanSet() {
			value, ok := os.LookupEnv(name)
			if !ok {
				return fmt.Errorf("env variable %s not set", name)
			}

			switch f.Kind() {
			case reflect.String:
				f.SetString(value)
			case reflect.Int:
				i, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("failed conversion for %s: %w", name, err)
				}
				f.SetInt(int64(i))
			case reflect.Float32, reflect.Float64:
				i, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("failed conversion for %s: %w", name, err)
				}
				f.SetFloat(i)
			case reflect.Bool:
				i, err := strconv.ParseBool(value)
				if err != nil {
					return fmt.Errorf("failed conversion for %s: %w", name, err)
				}
				f.SetBool(i)
			default:
				return fmt.Errorf("type %v not supported", f.Kind())
			}
		}
	}
	return nil
}

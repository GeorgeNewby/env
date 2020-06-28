package env_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GeorgeNewby/env"
)

func Example() {
	// .env file
	// HOST=localhost
	// PORT=8080

	// Errors ignored for brevity
	env.Load()

	config := struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}{}

	env.Unmarshal(&config)

	fmt.Printf("HOST: %s\n", config.Host)
	fmt.Printf("PORT: %d\n", config.Port)
	// Output: HOST: localhost
	// PORT: 8080
}

func TestLoad(t *testing.T) {
	if err := env.Load(); err != nil {
		log.Fatal(err)
	}

	host := os.Getenv("HOST")
	if host != "localhost" {
		t.Fatalf("HOST: expected localhost, got %s", host)
	}

	port := os.Getenv("PORT")
	if port != "8080" {
		t.Fatalf("PORT: expected 8080, got %s", port)
	}
}

func TestParse(t *testing.T) {
	tt := []struct {
		Name  string
		Input string
		Vars  map[string]string
	}{
		{
			Name:  "Empty file",
			Input: ``,
			Vars:  map[string]string{},
		},
		{
			Name:  "One variable",
			Input: `FOO=foo`,
			Vars: map[string]string{
				"FOO": "foo",
			},
		},
		{
			Name: "Multiple variables",
			Input: `
				FOO=foo
				BAR=bar
			`,
			Vars: map[string]string{
				"FOO": "foo",
				"BAR": "bar",
			},
		},
		{
			Name:  "Variable with gaps",
			Input: `  FOO  =  foo  `,
			Vars: map[string]string{
				"FOO": "foo",
			},
		},
		{
			Name: "Comments",
			Input: `
				# This is a comment
				FOO=foo
				#BAR=bar
			`,
			Vars: map[string]string{
				"FOO": "foo",
			},
		},
		{
			Name:  "Variable with space",
			Input: `FOO=foo bar`,
			Vars: map[string]string{
				"FOO": "foo bar",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := strings.NewReader(tc.Input)
			if err := env.Parse(r); err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			for k, v := range tc.Vars {
				env, ok := os.LookupEnv(k)
				if !ok {
					t.Fatalf("var %s not set, expected %s", k, v)
				}

				if v != env {
					t.Errorf("var %s expected %s, got %s", k, v, env)
				}
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		os.Setenv("STRING", "foo")
		os.Setenv("INTEGER", "5")
		os.Setenv("FLOAT_32", "51.432434")
		os.Setenv("FLOAT_64", "51.43243344285539")
		os.Setenv("BOOLEAN", "true")

		s := struct {
			String  string  `env:"STRING"`
			Integer int     `env:"INTEGER"`
			Float32 float32 `env:"FLOAT_32"`
			Float64 float32 `env:"FLOAT_64"`
			Boolean bool    `env:"BOOLEAN"`
		}{}
		if err := env.Unmarshal(&s); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if s.String != "foo" {
			t.Errorf("STRING: expected foo, got %v", s.String)
		}
		if s.Integer != 5 {
			t.Errorf("INTEGER: expected 5, got %v", s.Integer)
		}
		if s.Float32 != 51.432434 {
			t.Errorf("FLOAT_32: expected 51.432434, got %v", s.Float32)
		}
		if s.Float64 != 51.43243344285539 {
			t.Errorf("FLOAT_64: expected 51.43243344285539, got %v", s.Float64)
		}
		if s.Boolean != true {
			t.Errorf("BOOLEAN: expected true, got %v", s.Boolean)
		}
	})
}

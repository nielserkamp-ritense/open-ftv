package config

import (
	"flag"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

// TestRegisterFlags_Parity checks that RegisterFlags plus envflag.Parse fill each struct exactly as go-config does
// from its struct tags: the same defaults, and the same field for every environment variable.
//
// It only exists while both mechanisms coexist, and goes away together with go-config and the struct tags.
func TestRegisterFlags_Parity(t *testing.T) {
	fragments := []func() any{
		func() any { return &Server{} },
		func() any { return &InternalServer{} },
		func() any { return &HealthServer{} },
		func() any { return &Log{} },
		func() any { return &PDP{} },
		func() any { return &PIP{} },
		func() any { return &PAP{} },
		func() any { return &Cerbos{} },
		func() any { return &Authentication{} },
		func() any { return &Authorization{} },
		func() any { return &Bundle{} },
		func() any { return &OIDC{} },
	}

	const prefix = "PARITY_"

	load := func(t *testing.T, env map[string]string, newFragment func() any) (any, any) {
		t.Helper()

		for name, value := range env {
			t.Setenv(prefix+name, value)
		}

		want := newFragment()
		require.NoError(t, config.LoadConfig(want, config.WithArgs(), config.EnvironmentPrefix(prefix), config.NoHelpOnError()))

		got := newFragment()
		fs := flag.NewFlagSet("parity", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		got.(interface{ RegisterFlags(*flag.FlagSet) }).RegisterFlags(fs)

		lookup := func(name string) (string, bool) {
			value, ok := env[strings.TrimPrefix(name, prefix)]
			return value, ok
		}
		require.NoError(t, envflag.Parse(fs, nil, prefix, lookup))

		return want, got
	}

	for _, newFragment := range fragments {
		typ := reflect.TypeOf(newFragment()).Elem()

		t.Run(typ.Name()+"/defaults", func(t *testing.T) {
			want, got := load(t, nil, newFragment)
			assert.Equal(t, want, got)
		})

		for i := range typ.NumField() {
			field := typ.Field(i)

			names, ok := field.Tag.Lookup("env")
			if !ok {
				continue
			}

			// of the aliases (ADDRESS,HOST) only the last one remains, as that is the one deployments use.
			aliases := strings.Split(names, ",")
			name := aliases[len(aliases)-1]

			t.Run(typ.Name()+"/"+name, func(t *testing.T) {
				want, got := load(t, map[string]string{name: sample(field)}, newFragment)
				assert.Equal(t, want, got)
			})
		}
	}
}

// sample returns a valid value for the field that differs from its default.
func sample(field reflect.StructField) string {
	typ := field.Type

	switch {
	case typ == reflect.TypeOf(time.Duration(0)):
		return "7s"
	case typ.Kind() == reflect.Bool:
		if field.Tag.Get("default") == "true" {
			return "false"
		}
		return "true"
	case typ.Kind() == reflect.Int, typ.Kind() == reflect.Uint16:
		return "7"
	default:
		return "sample"
	}
}

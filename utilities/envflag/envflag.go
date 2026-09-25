// Package envflag parses a flag.FlagSet with environment variables as a fallback for flags not set on the command line.
package envflag

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// Parse parses args into fs. Every flag not set in args is then looked up in the environment as prefix plus the flag
// name in upper case, with '-' replaced by '_'; e.g. with prefix "PDP_", flag "log-level" is read from PDP_LOG_LEVEL.
//
// So a flag takes precedence over its environment variable, which takes precedence over the flag's default.
//
// A positional argument is an error: the flag package stops at the first one and would silently ignore every flag
// after it, as happens with "-log-source false -log-level debug", where "false" is not taken as the bool's value.
func Parse(fs *flag.FlagSet, args []string, prefix string, lookupEnv func(string) (string, bool)) error {
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q; a bool flag takes its value as -name=false", fs.Arg(0))
	}

	set := make(map[string]bool)

	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	var err error

	fs.VisitAll(func(f *flag.Flag) {
		if err != nil || set[f.Name] {
			return
		}

		name := envName(prefix, f.Name)
		if value, ok := lookupEnv(name); ok {
			if e := fs.Set(f.Name, value); e != nil {
				err = fmt.Errorf("invalid value %q for environment variable %s: %w", value, name, e)
			}
		}
	})

	return err
}

// envName returns the environment variable Parse reads for the given flag name.
func envName(prefix, flagName string) string {
	return prefix + strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}

// Uint16Var defines a uint16 flag, which the flag package lacks, e.g. for a port.
func Uint16Var(fs *flag.FlagSet, p *uint16, name string, value uint16, usage string) {
	*p = value
	fs.Var((*uint16Value)(p), name, usage)
}

type uint16Value uint16

func (v *uint16Value) Set(s string) error {
	n, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return err
	}

	*v = uint16Value(n)

	return nil
}

func (v *uint16Value) String() string { return strconv.FormatUint(uint64(*v), 10) }

package enumgen

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"
)

const Usage = `enumgen: turn a bag of constants into something a bit more useful

Usage: enumgen [options] <input>...

Inputs:

The <input> argument is a list of types contained in the current package, to which
methods will be added.

Methods generated:

- fmt.Stringer.String()
- IsValid() bool
- Lookup(str string) <T>
- encoding.TextMarshaler.MarshalText (with -textmarshal)
- encoding.TextUnmarshaler.UnmarshalText (with -textmarshal)
- flag.Value.Set(s string) (with -flagval)
`

type usageError string

func (u usageError) Error() string { return string(u) }

func IsUsageError(err error) bool {
	_, ok := err.(usageError)
	return ok
}

type Command struct {
	tags    string
	pkg     string
	flag    bool
	marshal bool
	format  bool
	out     string
}

func (cmd *Command) Flags(flags *flag.FlagSet) {
	flags.StringVar(&cmd.pkg, "pkg", os.Getenv("GOPACKAGE"), "package name to search for types (defaults to GOPACKAGE)")
	flags.StringVar(&cmd.out, "out", "enum_gen.go", "output file name")
	flags.StringVar(&cmd.tags, "tags", "", "comma-separated list of build tags")
	flags.BoolVar(&cmd.flag, "flag", true, "generate flag.Value")
	flags.BoolVar(&cmd.marshal, "marshal", false, "EXPERIMENTAL: generate encoding.TextMarshaler/TextUnmarshaler")
	flags.BoolVar(&cmd.format, "format", true, "run gofmt on result")
}

func (cmd *Command) Synopsis() string { return "Generate enum-ish helpers from a bag of constants" }
func (cmd *Command) Usage() string    { return Usage }

func (cmd *Command) Run(args ...string) error {
	if cmd.pkg == "" {
		return usageError("-pkg not set")
	}
	if cmd.out == "" {
		return usageError("-out not set")
	}

	tags := strings.Split(cmd.tags, ",")

	g := &generator{
		withFlagVal: cmd.flag,
		withMarshal: cmd.marshal,
		format:      cmd.format,
	}

	pkg, err := g.parsePackage(cmd.pkg, tags)
	if err != nil {
		return err
	}

	for _, typeName := range args {
		cns, err := g.extract(pkg, typeName)
		if err != nil {
			return err
		}

		if err := g.generate(cns); err != nil {
			return err
		}
	}

	out, err := g.Output(cmd.out, pkg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(cmd.out, out, 0644)
}
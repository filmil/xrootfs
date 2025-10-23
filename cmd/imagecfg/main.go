package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
)

type RepeatStr struct {
	values []string
}

var _ flag.Value = (*RepeatStr)(nil)

func (self *RepeatStr) String() string {
	return strings.Join(self.values, ";")
}

func (self *RepeatStr) Set(value string) error {
	self.values = append(self.values, value)
	return nil
}

func (self *RepeatStr) Values() []string {
	return self.values
}

// AsKeyvals returns the repeated flag as a map of keyvalues.
func (self *RepeatStr) AsKeyvals() (*map[string]string, error) {
	ret := map[string]string{}
	for _, value := range self.values {
		keyVals := strings.SplitN(value, "=", 2)
		if len(keyVals) != 2 {
			return nil, fmt.Errorf("invalid format: %q\n\texpected key=value, was: %+v", value, keyVals)
		}
		key := keyVals[0]
		vals := keyVals[1]
		ret[key] = vals
	}
	return &ret, nil
}

// AsMap returns the RepeatStr as a map of string to string slices.
func (self *RepeatStr) AsMap() (*map[string][]string, error) {
	ret := map[string][]string{}

	for _, value := range self.values {
		keyVals := strings.SplitN(value, "=", 2)
		if len(keyVals) != 2 {
			return nil, fmt.Errorf("invalid format: %q, expected key=value, was: %+v", value, keyVals)
		}
		key := keyVals[0]
		vals := strings.Split(keyVals[1], ",")
		ret[key] = append(ret[key], vals...)
	}
	return &ret, nil
}

func main() {
	log.SetPrefix(fmt.Sprintf("%s: ", os.Args[0]))

	var (
		packages                 RepeatStr
		archs                    RepeatStr
		sources                  RepeatStr
		templateName, outputName string
	)
	flag.Var(&packages, "package", "A package to include, such as `cpio`")
	flag.Var(&archs, "arch", "An arch to include, such as `amd64`")
	flag.Var(&sources, "source", "Map from URL to comma separated list of channels, such as --source=https://snapshot.ubuntu.com/=noble,main,restricted,universe,multiverse")
	flag.StringVar(&templateName, "template", "", "template")
	flag.StringVar(&outputName, "output", "", "templateName")
	flag.Parse()

	if templateName == "" {
		log.Printf("flag --template=... is required")
		os.Exit(1)
	}
	if outputName == "" {
		log.Printf("flag --output=... is required")
		os.Exit(1)
	}
	f, err := os.Open(templateName)
	if err != nil {
		log.Printf("could not open template: %q: %w", templateName, err)
		os.Exit(1)
	}

	o, err := os.Create(outputName)
	if err != nil {
		log.Printf("could not create output: %q: %v", err, outputName)
	}
	defer o.Close()

	if err := run(packages, archs, sources, f, o); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

type Source struct {
	Channels []string
	URL      string
}

type RootFsValues struct {
	Archs    []string
	Sources  []Source
	Packages []string
}

func run(packages, archs, sources RepeatStr, r io.Reader, output io.Writer) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read from input: %w", err)
	}
	s := string(b)
	t, err := template.New("tpl").Parse(s)
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var data RootFsValues
	data.Archs = archs.Values()
	data.Packages = packages.Values()
	srcs, err := sources.AsMap()
	if err != nil {
		return fmt.Errorf("could not parse sources:\n\t%+v: %w", sources, err)
	}
	for k, vs := range *srcs {
		var s Source
		s.URL = k
		s.Channels = vs
		data.Sources = append(data.Sources, s)
	}

	if err := t.Execute(output, data); err != nil {
		return fmt.Errorf("could not apply data to template: %v: %w", data, err)
	}

	return nil
}

package main

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var (
	tplName string
)

func init() {
	flag.StringVar(&tplName, "template", "", "")
	flag.Parse()

	if tplName == "" {
		panic("flag --tplName is required")
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name                     string
		packages, archs, sources RepeatStr
		expected                 string
	}{
		{
			name:     "basic",
			packages: RepeatStr{values: []string{"tar", "cpio"}},
			archs:    RepeatStr{values: []string{"arm64"}},
			sources:  RepeatStr{values: []string{"https://www.foo=main,bar"}},
			expected: `version: 1
sources:
  - channel: main bar
    url: https://www.foo
archs:
  - arm64
packages:
  - tar
  - cpio

`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := os.Open(tplName)
			if err != nil {
				t.Fatalf("could not find template to open: %v", err)
			}

			var buf strings.Builder
			if err := run(test.packages, test.archs, test.sources, f, &buf); err != nil {
				t.Errorf("run errored out: %v", err)
			}
			if test.expected != buf.String() {
				t.Errorf("want:\n\"%v\"\n\tgot:\n\"%v\"", test.expected, buf.String())
			}
		})
	}
}

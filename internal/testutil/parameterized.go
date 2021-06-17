package testutil

import (
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

// RunParameterizedTests runs parameterized tests.
func RunParameterizedTests(r Reporter, e ParameterizedTestExecutor, files ...string) {
	r.Helper()
	for _, file := range files {
		file := file
		run(r, file, func(r Reporter) {
			f, err := os.Open(file)
			if err != nil {
				r.Fatalf("failed to open file: %s", err)
			}
			dec := yaml.NewDecoder(f, yaml.UseOrderedMap())
			for {
				var p TestParameter
				if err := dec.Decode(&p); err != nil {
					if err == io.EOF {
						break
					} else {
						r.Fatalf("failed to decode test parameter: %s", err)
					}
				}
				run(r, p.Name, func(r Reporter) {
					exec := e(r, func(v interface{}) {
						b, err := yaml.Marshal(p.YAML)
						if err != nil {
							r.Fatalf("failed to marshal YAML: %s", err)
						}
						if err := yaml.UnmarshalWithOptions(b, v, yaml.UseOrderedMap()); err != nil {
							r.Fatalf("failed to unmarshal YAML: %s", err)
						}
					})
					run(r, "OK", func(r Reporter) {
						for i, ok := range p.OKs {
							ok := ok
							run(r, fmt.Sprint(i), func(r Reporter) {
								err := exec(r, ok)
								if err != nil {
									r.Fatal(err)
								}
							})
						}
					})
					run(r, "NG", func(r Reporter) {
						for i, ng := range p.NGs {
							ng := ng
							run(r, fmt.Sprint(i), func(r Reporter) {
								err := exec(r, ng)
								if err == nil {
									r.Fatal("no error")
								}
							})
						}
					})
				})
			}
		})
	}
}

// TestParameter is a parameters for parameterized testing.
type TestParameter struct {
	Name string        `yaml:"name"`
	YAML interface{}   `yaml:"yaml"`
	OKs  []interface{} `yaml:"ok"`
	NGs  []interface{} `yaml:"ng"`
}

// ParameterizedTestExecutor represents a executor for parameterized testing.
type ParameterizedTestExecutor func(Reporter, func(interface{})) func(Reporter, interface{}) error

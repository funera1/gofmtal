package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	flagUpdate bool
)

func init() {
	flag.BoolVar(&flagUpdate, "update", false, "update golden files")
}

func Test(t *testing.T) {
	// determine input files
	match, err := filepath.Glob("../testdata/*.input")
	if err != nil {
		t.Fatal(err)
	}

	for _, in := range match {
		name := filepath.Base(in)
		t.Run(name, func(t *testing.T) {
			out := in // for files where input and output are identical
			if strings.HasSuffix(in, ".input") {
				out = in[:len(in)-len(".input")] + ".golden"
			}

			got, err := processFile(in)
			if err != nil {
				t.Error(err)
				return
			}

			want, err := os.ReadFile(out)
			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

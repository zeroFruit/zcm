package oci

import (
	"bufio"
	"bytes"

	"github.com/opencontainers/runtime-tools/generate"
)

type RuntimeSpec []byte

type SpecOptions struct {
	Command      []string
	Args         []string
	RootPath     string
	RootReadonly bool
}

func NewSpec(opts SpecOptions) (RuntimeSpec, error) {
	gen, err := generate.New("linux")
	if err != nil {
		return nil, err
	}
	gen.HostSpecific = true
	gen.SetRootPath(opts.RootPath)
	gen.SetRootReadonly(opts.RootReadonly)
	gen.SetProcessArgs(append(opts.Command, opts.Args...))

	var buf bytes.Buffer
	exportOpts := generate.ExportOptions{}
	if err := gen.Save(bufio.NewWriter(&buf), exportOpts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

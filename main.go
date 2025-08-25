package main

import (
	"fmt"
	"terraform-config-parser/pkg/parser"

	"github.com/spf13/afero"
)

func main() {
	fs := afero.Afero{Fs: afero.OsFs{}}

	p := parser.NewParser(&fs)
	tfconfig, err := p.ParseTerraformWorkspace("sample")
	if err != nil {
		panic(err)
	}

	summary, err := tfconfig.Summary(true)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(summary))
}

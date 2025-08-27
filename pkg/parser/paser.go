package parser

import (
	"errors"
	"path/filepath"
	"terraform-config-parser/pkg/filesystem"
	"terraform-config-parser/pkg/parser/schema"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Parser struct {
	fs  filesystem.FileReader
	hcl *hclparse.Parser
}

func NewParser(fs filesystem.FileReader) *Parser {
	return &Parser{
		fs:  fs,
		hcl: hclparse.NewParser(),
	}
}

func (p *Parser) ParseTerraformWorkspace(dir string) (*TerraformConfig, error) {
	exist, err := p.fs.DirExists(dir)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.New("directory does not exist")
	}

	dirFiles, err := p.fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	aggBlocks := []schema.Block{}

	for _, dirFile := range dirFiles {
		if dirFile.IsDir() || filepath.Ext(dirFile.Name()) != ".tf" {
			continue
		}

		hclFile, err := p.loadHcl(filepath.Join(dir, dirFile.Name()))
		if err != nil {
			return nil, err
		}

		blocks, err := p.parseBlocks(hclFile)
		if err != nil {
			return nil, err
		}

		aggBlocks = append(aggBlocks, blocks...)
	}

	return generateTerraformConfig(aggBlocks), nil
}

func (p *Parser) loadHcl(filename string) (*hcl.File, error) {
	content, err := p.fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	file, diags := p.hcl.ParseHCL(content, filename)
	if file == nil || file.Body == nil || diags.HasErrors() {
		return nil, errors.Join(diags.Errs()...)
	}

	return file, nil
}

func (p *Parser) parseBlocks(file *hcl.File) ([]schema.Block, error) {
	rootBody := file.Body.(*hclsyntax.Body)

	blocks := []schema.Block{}
	for _, block := range rootBody.Blocks {
		var parsedBlock schema.Block = nil

		switch block.Type {
		case "variable":
			parsedBlock = &schema.Variable{}
		case "output":
			parsedBlock = &schema.Output{}
		case "terraform":
			parsedBlock = &schema.Terraform{}
		default:
			continue
		}

		if err := parsedBlock.Parse(file, block); err != nil {
			return nil, err
		}

		blocks = append(blocks, parsedBlock)
	}

	return blocks, nil
}

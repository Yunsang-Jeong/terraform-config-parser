package parser

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/filesystem"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/parser/schema"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Mode int

const (
	Simple Mode = iota
	Detail
)

type Parser struct {
	fs   filesystem.FileReader
	hcl  *hclparse.Parser
	mode Mode
}

func NewParser(fs filesystem.FileReader, mode Mode) *Parser {
	return &Parser{
		fs:   fs,
		hcl:  hclparse.NewParser(),
		mode: mode,
	}
}


func (p *Parser) ParseTerraformWorkspace(dir string) (*TerraformConfig, error) {
	logger.InfoKV("Starting terraform workspace parsing", "directory", dir)

	exist, err := p.fs.DirExists(dir)
	if err != nil {
		logger.ErrorKV("Failed to check terraform workspace directory", "directory", dir, "error", err)
		return nil, fmt.Errorf("failed to check terraform workspace directory: %w", err)
	}
	if !exist {
		logger.ErrorKV("Terraform workspace directory not found", "directory", dir)
		return nil, fmt.Errorf("terraform workspace directory not found: %s", dir)
	}

	dirFiles, err := p.fs.ReadDir(dir)
	if err != nil {
		logger.ErrorKV("Failed to read terraform workspace directory", "directory", dir, "error", err)
		return nil, fmt.Errorf("failed to read terraform workspace directory %s: %w", dir, err)
	}

	logger.DebugKV("Found files in directory", "directory", dir, "file_count", len(dirFiles))

	aggBlocks := []schema.Block{}

	for _, dirFile := range dirFiles {
		if dirFile.IsDir() || filepath.Ext(dirFile.Name()) != ".tf" {
			logger.DebugKV("Skipping non-terraform file", "file", dirFile.Name())
			continue
		}

		logger.DebugKV("Processing terraform file", "file", dirFile.Name())

		hclFile, err := p.loadHcl(filepath.Join(dir, dirFile.Name()))
		if err != nil {
			logger.ErrorKV("Failed to load terraform file", "directory", dir, "file", dirFile.Name(), "error", err)
			return nil, fmt.Errorf("failed to load terraform file %s: %w", dirFile.Name(), err)
		}

		blocks, err := p.parseBlocks(hclFile)
		if err != nil {
			logger.ErrorKV("Failed to parse terraform blocks", "directory", dir, "file", dirFile.Name(), "mode", p.getModeString(), "error", err)
			return nil, fmt.Errorf("failed to parse terraform blocks in %s: %w", dirFile.Name(), err)
		}

		logger.DebugKV("Successfully parsed blocks", "directory", dir, "file", dirFile.Name(), "block_count", len(blocks), "mode", p.getModeString())
		aggBlocks = append(aggBlocks, blocks...)
	}

	tfConfig := generateTerraformConfig(aggBlocks)
	logger.InfoKV("Successfully parsed terraform workspace",
		"directory", dir,
		"variables", len(tfConfig.Variables),
		"outputs", len(tfConfig.Outputs),
		"terraform_blocks", len(tfConfig.Terraform))

	return tfConfig, nil
}

func (p *Parser) loadHcl(filename string) (*hcl.File, error) {
	content, err := p.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read terraform file %s: %w", filename, err)
	}

	file, diags := p.hcl.ParseHCL(content, filename)
	if file == nil || file.Body == nil || diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL syntax in %s: %w", filename, errors.Join(diags.Errs()...))
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

		case "resource", "data", "module", "provider", "locals":
			if p.mode != Detail {
				continue
			}
			// TODO: Implement parsing for these block types when needed
			continue

		default:
			continue
		}

		if err := parsedBlock.Parse(file, block); err != nil {
			return nil, fmt.Errorf("failed to parse %s block: %w", block.Type, err)
		}

		blocks = append(blocks, parsedBlock)
	}

	return blocks, nil
}

func (p *Parser) getModeString() string {
	switch p.mode {
	case Simple:
		return "Simple"
	case Detail:
		return "Detail"
	default:
		return "Unknown"
	}
}

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
	"go.uber.org/zap"
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
	logger.Info("Starting terraform workspace parsing", zap.String("directory", dir))

	exist, err := p.fs.DirExists(dir)
	if err != nil {
		logger.Error("Failed to check terraform workspace directory", zap.String("directory", dir), zap.Error(err))
		return nil, fmt.Errorf("failed to check terraform workspace directory: %w", err)
	}
	if !exist {
		logger.Error("Terraform workspace directory not found", zap.String("directory", dir))
		return nil, fmt.Errorf("terraform workspace directory not found: %s", dir)
	}

	dirFiles, err := p.fs.ReadDir(dir)
	if err != nil {
		logger.Error("Failed to read terraform workspace directory", zap.String("directory", dir), zap.Error(err))
		return nil, fmt.Errorf("failed to read terraform workspace directory %s: %w", dir, err)
	}

	logger.Debug("Found files in directory", zap.String("directory", dir), zap.Int("file_count", len(dirFiles)))

	aggBlocks := []schema.Block{}

	for _, dirFile := range dirFiles {
		if dirFile.IsDir() || filepath.Ext(dirFile.Name()) != ".tf" {
			logger.Debug("Skipping non-terraform file", zap.String("file", dirFile.Name()))
			continue
		}

		logger.Debug("Processing terraform file", zap.String("file", dirFile.Name()))

		hclFile, err := p.loadHcl(filepath.Join(dir, dirFile.Name()))
		if err != nil {
			logger.Error("Failed to load terraform file", zap.String("file", dirFile.Name()), zap.Error(err))
			return nil, fmt.Errorf("failed to load terraform file %s: %w", dirFile.Name(), err)
		}

		blocks, err := p.parseBlocks(hclFile)
		if err != nil {
			logger.Error("Failed to parse terraform blocks", zap.String("file", dirFile.Name()), zap.Error(err))
			return nil, fmt.Errorf("failed to parse terraform blocks in %s: %w", dirFile.Name(), err)
		}

		logger.Debug("Successfully parsed blocks", zap.String("file", dirFile.Name()), zap.Int("block_count", len(blocks)))
		aggBlocks = append(aggBlocks, blocks...)
	}

	tfConfig := generateTerraformConfig(aggBlocks)
	logger.Info("Successfully parsed terraform workspace",
		zap.String("directory", dir),
		zap.Int("variables", len(tfConfig.Variables)),
		zap.Int("outputs", len(tfConfig.Outputs)),
		zap.Int("terraform_blocks", len(tfConfig.Terraform)))

	return tfConfig, nil
}

func (p *Parser) loadHcl(filename string) (*hcl.File, error) {
	logger.Debug("Loading HCL file", zap.String("filename", filename))

	content, err := p.fs.ReadFile(filename)
	if err != nil {
		logger.Error("Failed to read terraform file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to read terraform file %s: %w", filename, err)
	}

	file, diags := p.hcl.ParseHCL(content, filename)
	if file == nil || file.Body == nil || diags.HasErrors() {
		logger.Error("Failed to parse HCL syntax", zap.String("filename", filename), zap.String("diagnostics", diags.Error()))
		return nil, fmt.Errorf("failed to parse HCL syntax in %s: %w", filename, errors.Join(diags.Errs()...))
	}

	logger.Debug("Successfully loaded HCL file", zap.String("filename", filename))

	return file, nil
}

func (p *Parser) parseBlocks(file *hcl.File) ([]schema.Block, error) {
	rootBody := file.Body.(*hclsyntax.Body)
	logger.Debug("Parsing blocks from HCL file", zap.Int("block_count", len(rootBody.Blocks)))

	blocks := []schema.Block{}
	for _, block := range rootBody.Blocks {
		var parsedBlock schema.Block = nil

		logger.Debug("Processing block", zap.String("type", block.Type), zap.Strings("labels", block.Labels))

		switch block.Type {
		case "variable":
			parsedBlock = &schema.Variable{}
		case "output":
			parsedBlock = &schema.Output{}
		case "terraform":
			parsedBlock = &schema.Terraform{}
		default:
			logger.Debug("Skipping unsupported block type", zap.String("type", block.Type))
			continue
		}

		if err := parsedBlock.Parse(file, block); err != nil {
			logger.Error("Failed to parse block", zap.String("type", block.Type), zap.Strings("labels", block.Labels), zap.Error(err))
			return nil, fmt.Errorf("failed to parse %s block: %w", block.Type, err)
		}

		logger.Debug("Successfully parsed block", zap.String("type", block.Type), zap.Strings("labels", block.Labels))

		blocks = append(blocks, parsedBlock)
	}

	return blocks, nil
}

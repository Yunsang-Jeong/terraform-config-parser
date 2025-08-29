package parser

import (
	"bytes"
	"encoding/json"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/parser/schema"
)

type TerraformConfig struct {
	Variables []*schema.Variable  `json:"variables,omitempty"`
	Outputs   []*schema.Output    `json:"outputs,omitempty"`
	Terraform []*schema.Terraform `json:"terraform,omitempty"`
}

func generateTerraformConfig(blocks []schema.Block) *TerraformConfig {
	tfconfig := TerraformConfig{
		Variables: make([]*schema.Variable, 0),
		Outputs:   make([]*schema.Output, 0),
		Terraform: make([]*schema.Terraform, 0),
	}

	for _, block := range blocks {
		switch b := block.(type) {
		case *schema.Variable:
			tfconfig.Variables = append(tfconfig.Variables, b)
		case *schema.Output:
			tfconfig.Outputs = append(tfconfig.Outputs, b)
		case *schema.Terraform:
			tfconfig.Terraform = append(tfconfig.Terraform, b)
		}
	}

	return &tfconfig
}

func (t *TerraformConfig) Summary(pretty bool) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(&t); err != nil {
		return nil, err
	}

	return bytes.TrimSpace(buf.Bytes()), nil
}

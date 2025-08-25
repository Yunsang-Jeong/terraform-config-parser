package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Terraform struct {
	RequiredVersion   string                        `json:"required_version,omitempty"`
	Experiments       []string                      `json:"experiments,omitempty"`
	RequiredProviders map[string]*RequiredProvider  `json:"required_providers,omitempty"`
}

type RequiredProvider struct {
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
}

func (b *Terraform) Parse(file *hcl.File, block *hclsyntax.Block) error {
	if len(block.Labels) != 0 {
		return fmt.Errorf("terraform block must not have labels")
	}

	attrs := block.Body.Attributes

	if requiredVersionAttr, ok := attrs["required_version"]; ok {
		b.RequiredVersion = parseAttributeToString(file, requiredVersionAttr)
	}

	if experimentsAttr, ok := attrs["experiments"]; ok {
		b.Experiments = parseAttributeToStringList(file, experimentsAttr)
	}

	b.RequiredProviders = make(map[string]*RequiredProvider)
	for _, blockInBlock := range block.Body.Blocks {
		switch blockInBlock.Type {
		case "required_providers":
			// required_providers 블록 내의 각 provider 파싱
			for providerName, attr := range blockInBlock.Body.Attributes {
				// 범용 함수로 객체를 맵으로 파싱
				providerConfig := parseAttributeToStringMap(file, attr)
				
				provider := &RequiredProvider{
					Source:  providerConfig["source"],
					Version: providerConfig["version"],
				}
				
				b.RequiredProviders[providerName] = provider
			}
		}
	}

	return nil
}


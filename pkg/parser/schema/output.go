package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Output struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	// Value       string `json:"value"`
}

func (b *Output) Parse(file *hcl.File, block *hclsyntax.Block) error {
	if len(block.Labels) != 1 {
		return fmt.Errorf("variable block must have one label")
	}
	b.Name = block.Labels[0]

	attrs := block.Body.Attributes

	// if valueAttr, ok := attrs["value"]; ok {
	// 	b.Value = parseAttributeToString(file, valueAttr)
	// } else {
	// 	return fmt.Errorf("variable %s is missing Value attribute", b.Name)
	// }

	if descriptionAttr, ok := attrs["description"]; ok {
		b.Description = parseAttributeToString(file, descriptionAttr)
	}

	if sensitiveAttr, ok := attrs["sensitive"]; ok {
		b.Sensitive = parseAttributeToBool(file, sensitiveAttr)
	}

	return nil
}

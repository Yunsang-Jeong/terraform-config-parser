package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Variable struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Type        string                `json:"type,omitempty"`
	Default     interface{}           `json:"default,omitempty"`
	Required    bool                  `json:"required"`
	Sensitive   bool                  `json:"sensitive"`
	Validation  []*VariableValidation `json:"validation,omitempty"`
}

type VariableValidation struct {
	Condition    string `json:"condition"`
	ErrorMessage string `json:"error_message"`
}

func (b *Variable) Parse(file *hcl.File, block *hclsyntax.Block) error {
	if len(block.Labels) != 1 {
		return fmt.Errorf("variable block must have one label")
	}
	b.Name = block.Labels[0]

	attrs := block.Body.Attributes

	if descAttr, ok := attrs["description"]; ok {
		b.Description = parseAttributeToString(file, descAttr)
	}

	if typeAttr, ok := attrs["type"]; ok {
		b.Type = parseAttributeToString(file, typeAttr)
	}

	if defaultAttr, ok := attrs["default"]; ok {
		b.Default = parseAttributeToInterface(file, defaultAttr)
	} else {
		b.Required = true
	}

	if sensitiveAttr, ok := attrs["sensitive"]; ok {
		b.Sensitive = parseAttributeToBool(file, sensitiveAttr)
	}

	for _, blockInBlock := range block.Body.Blocks {
		switch blockInBlock.Type {
		case "validation":
			validation := &VariableValidation{}
			if err := validation.Parse(file, blockInBlock); err != nil {
				return fmt.Errorf("error parsing validation for variable %s: %w", b.Name, err)
			}

			b.Validation = append(b.Validation, validation)
		}
	}

	return nil
}

func (b *VariableValidation) Parse(file *hcl.File, block *hclsyntax.Block) error {
	attrs := block.Body.Attributes

	if conditionAttr, ok := attrs["condition"]; ok {
		b.Condition = parseAttributeToString(file, conditionAttr)
	} else {
		return fmt.Errorf("condition is missing in validation block")
	}

	if errorMessageAttr, ok := attrs["error_message"]; ok {
		b.ErrorMessage = parseAttributeToString(file, errorMessageAttr)
	} else {
		return fmt.Errorf("error_message is missing in validation block")
	}

	return nil
}

package schema

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

/*
HCL/Terraform Expression Types Overview:

## Basic Scalar Types (LiteralValueExpr):
- Numbers: 42, 3.14, 1e10
- Booleans: true, false
- null: null

## String-related:
- Template expressions (TemplateExpr): "hello", "prefix-${var.name}-suffix"
  └─ Simple strings: "hello" -> TemplateExpr.Parts[0] is LiteralValueExpr
  └─ Interpolated strings: "hello ${var.name}" -> Parts contains LiteralValueExpr + TemplateWrapExpr

## Variable references and namespace access:
- Scope traversal (ScopeTraversalExpr): All expressions that access values using dot notation
  └─ Simple names: source, version (unquoted single names)
  └─ Variable references: var.name, local.value
  └─ Resource references: aws_instance.web.id, data.aws_vpc.main.cidr_block
  └─ Module references: module.network.vpc_id

## Complex structures:
- Tuples/arrays (TupleConsExpr): [1, 2, 3], ["a", "b"]
- Objects/maps (ObjectConsExpr): {key = "value", foo = "bar"}
  └─ Key expressions (ObjectConsKeyExpr): Wrapper for object keys
      └─ When Wrapped is ScopeTraversalExpr: source = "value"
      └─ When Wrapped is TemplateExpr: "source" = "value"

## Functions and operations:
- Function calls (FunctionCallExpr): toset([1,2,3]), length(var.list)
- Conditional expressions (ConditionalExpr): condition ? true_val : false_val
- Binary operations (BinaryOpExpr): a + b, a == b
- Unary operations (UnaryOpExpr): !condition, -number

## Advanced expressions:
- Index access (IndexExpr): var.list[0], var.map["key"]
- Attribute access (GetAttrExpr): object.attribute
- Splat expressions (SplatExpr): var.list[*].name
- For expressions (ForExpr): [for item in list : item.name]
- Parentheses expressions (ParenthesesExpr): (expression)

## Examples:
variable "example" {
  type = string              # <- ScopeTraversalExpr (simple name 'string')
  default = "hello"          # <- TemplateExpr (quoted string)
  count = 42                 # <- LiteralValueExpr (number)
  enabled = true             # <- LiteralValueExpr (boolean)
  tags = {                   # <- ObjectConsExpr
    Name = "test"            # <- Name: ScopeTraversalExpr (simple name), "test": TemplateExpr
    "Environment" = "dev"    # <- "Environment": TemplateExpr, "dev": TemplateExpr
  }
  computed = var.prefix      # <- ScopeTraversalExpr (variable reference)
}
*/

type Block interface {
	Parse(file *hcl.File, block *hclsyntax.Block) error
}

func parseAttributeToInterface(file *hcl.File, attr *hclsyntax.Attribute) interface{} {
	//
	// Return literal string, number, bool, null values with their proper types
	// For example, 'string' in a variable block's type is LiteralValue
	// while "default" in default is a Template
	//
	if lv, ok := attr.Expr.(*hclsyntax.LiteralValueExpr); ok {
		switch lv.Val.Type() {
		case cty.String:
			// Convert HCL string to Go string
			return lv.Val.AsString()
		case cty.Number:
			if lv.Val.AsBigFloat().IsInt() {
				// Convert HCL integer to Go int64
				i, _ := lv.Val.AsBigFloat().Int64()
				return i
			}

			// Convert HCL non-integer number to Go float64
			f, _ := lv.Val.AsBigFloat().Float64()
			return f
		case cty.Bool:
			// Convert HCL bool to Go bool
			return lv.Val.True()
		}
		if lv.Val.IsNull() {
			// Convert HCL null to Go nil
			return nil
		}
	}

	// Extract string from template expressions
	if te, ok := attr.Expr.(*hclsyntax.TemplateExpr); ok {
		if len(te.Parts) == 1 {
			if lv, ok := te.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lv.Val.Type() == cty.String {
				return lv.Val.AsString()
			}
		}
	}

	// For complex expressions, return original HCL syntax
	raw := attr.Expr.Range().SliceBytes(file.Bytes)
	return strings.TrimSpace(string(raw))
}

func parseAttributeToString(file *hcl.File, attr *hclsyntax.Attribute) string {
	value := parseAttributeToInterface(file, attr)
	if str, ok := value.(string); ok {
		return str
	}

	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseAttributeToBool(file *hcl.File, attr *hclsyntax.Attribute) bool {
	value := parseAttributeToInterface(file, attr)
	if boolVal, ok := value.(bool); ok {
		return boolVal
	}

	switch v := value.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true":
			return true
		case "false":
			return false
		}
	default:
		return false
	}

	return false
}

// Parse array attributes to string slice
func parseAttributeToStringList(file *hcl.File, attr *hclsyntax.Attribute) []string {
	// Handle tuple expressions (HCL arrays)
	if tupleExpr, ok := attr.Expr.(*hclsyntax.TupleConsExpr); ok {
		result := make([]string, 0, len(tupleExpr.Exprs))
		for _, expr := range tupleExpr.Exprs {
			fakeAttr := &hclsyntax.Attribute{Expr: expr}
			result = append(result, parseAttributeToString(file, fakeAttr))
		}
		return result
	}

	// For complex expressions, parse original simply
	raw := string(attr.Expr.Range().SliceBytes(file.Bytes))
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "[]")

	if raw == "" {
		return []string{}
	}

	// Simple comma splitting (can be improved for more sophisticated parsing)
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// Parse object attributes to map (key-value pairs)
func parseAttributeToStringMap(file *hcl.File, attr *hclsyntax.Attribute) map[string]string {
	result := make(map[string]string)

	// Handle object expressions
	if objExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range objExpr.Items {
			key := extractObjectKey(item.KeyExpr)
			if key != "" {
				fakeAttr := &hclsyntax.Attribute{Expr: item.ValueExpr}
				value := parseAttributeToString(file, fakeAttr)
				result[key] = value
			}
		}
	}

	return result
}

// Helper function to extract object keys
func extractObjectKey(keyExpr hclsyntax.Expression) string {
	switch key := keyExpr.(type) {
	case *hclsyntax.ObjectConsKeyExpr:
		if key.Wrapped != nil {
			if literalKey, ok := key.Wrapped.(*hclsyntax.LiteralValueExpr); ok {
				return literalKey.Val.AsString()
			}
			if scopeKey, ok := key.Wrapped.(*hclsyntax.ScopeTraversalExpr); ok {
				return scopeKey.Traversal.RootName()
			}
		}
		return ""
	case *hclsyntax.LiteralValueExpr:
		return key.Val.AsString()
	case *hclsyntax.ScopeTraversalExpr:
		// For identifiers (e.g., source, version)
		if len(key.Traversal) > 0 {
			if rootName := key.Traversal.RootName(); rootName != "" {
				return rootName
			}
		}
	}
	return ""
}

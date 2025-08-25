# terraform-config-parser

Parse and report terraform workspace (config files).

## Supported Terraform Constructs

The parser currently supports:

### Variable Blocks
- All Terraform types: `string`, `number`, `bool`, `list()`, `map()`, `object()`, `tuple()`, `set()`, `any`
- Variable attributes: `type`, `description`, `default`, `sensitive`, `nullable`, `validation`
- Complex default values and validation rules

### Output Blocks
- Output value expressions
- Output descriptions and sensitive flags

### Terraform Blocks
- Terraform configuration settings
- Required providers and versions

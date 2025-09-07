package parser

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/filesystem"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/parser/schema"
)

// testFileSystem wraps fstest.MapFS to implement filesystem.FileReader interface
type testFileSystem struct {
	mapFS fstest.MapFS
}

func (tfs *testFileSystem) DirExists(dirname string) (bool, error) {
	dirname = strings.TrimPrefix(dirname, "./")
	if dirname == "" || dirname == "." {
		return true, nil
	}

	for path := range tfs.mapFS {
		if strings.HasPrefix(path, dirname+"/") || path == dirname {
			return true, nil
		}
	}
	return false, nil
}

func (tfs *testFileSystem) ReadDir(dirname string) ([]os.FileInfo, error) {
	dirname = strings.TrimPrefix(dirname, "./")
	if dirname == "" {
		dirname = "."
	}

	entries, err := fs.ReadDir(tfs.mapFS, dirname)
	if err != nil {
		return nil, err
	}

	var fileInfos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	return fileInfos, nil
}

func (tfs *testFileSystem) ReadFile(filename string) ([]byte, error) {
	filename = strings.TrimPrefix(filename, "./")
	return fs.ReadFile(tfs.mapFS, filename)
}

func newTestFileSystem(files map[string]string) filesystem.FileReader {
	mapFS := fstest.MapFS{}
	for filename, content := range files {
		mapFS[filename] = &fstest.MapFile{
			Data: []byte(content),
		}
	}
	return &testFileSystem{mapFS: mapFS}
}

// Test expectations structure
type TestExpectations struct {
	VariableCount     *int
	OutputCount       *int
	TerraformCount    *int
	Variables         map[string]*VariableExpectation
	Outputs           map[string]*OutputExpectation
	TerraformSettings *TerraformExpectation
}

type VariableExpectation struct {
	Type            *string
	HasDefault      *bool
	Sensitive       *bool
	Required        *bool
	ValidationCount *int
}

type OutputExpectation struct {
	Sensitive *bool
}

type TerraformExpectation struct {
	RequiredVersion *string
	ProviderCount   *int
	ExperimentCount *int
	Providers       map[string]*ProviderExpectation
}

type ProviderExpectation struct {
	Source  *string
	Version *string
}

// Generic validation helper functions
func validateCount[T any](t *testing.T, items []T, expected int, itemType string) {
	t.Helper()
	if len(items) != expected {
		t.Errorf("Expected %d %s, got %d", expected, itemType, len(items))
	}
}

func validateExpectations(t *testing.T, config *TerraformConfig, expectations TestExpectations) {
	t.Helper()

	// Validate counts
	if expectations.VariableCount != nil {
		validateCount(t, config.Variables, *expectations.VariableCount, "variables")
	}
	if expectations.OutputCount != nil {
		validateCount(t, config.Outputs, *expectations.OutputCount, "outputs")
	}
	if expectations.TerraformCount != nil {
		validateCount(t, config.Terraform, *expectations.TerraformCount, "terraform blocks")
	}

	// Validate specific variables
	for name, expectation := range expectations.Variables {
		variable := findVariable(t, config, name)
		if variable != nil {
			validateVariableExpectation(t, variable, expectation)
		}
	}

	// Validate specific outputs
	for name, expectation := range expectations.Outputs {
		output := findOutput(t, config, name)
		if output != nil {
			validateOutputExpectation(t, output, expectation)
		}
	}

	// Validate terraform settings
	if expectations.TerraformSettings != nil {
		validateTerraformExpectation(t, config, expectations.TerraformSettings)
	}
}

func findVariable(t *testing.T, config *TerraformConfig, name string) *schema.Variable {
	t.Helper()
	for _, v := range config.Variables {
		if v.Name == name {
			return v
		}
	}
	t.Errorf("Variable %s not found", name)
	return nil
}

func findOutput(t *testing.T, config *TerraformConfig, name string) *schema.Output {
	t.Helper()
	for _, o := range config.Outputs {
		if o.Name == name {
			return o
		}
	}
	t.Errorf("Output %s not found", name)
	return nil
}

func validateVariableExpectation(t *testing.T, variable *schema.Variable, expectation *VariableExpectation) {
	t.Helper()
	if expectation.Type != nil && variable.Type != *expectation.Type {
		t.Errorf("Variable %s: expected type %s, got %s", variable.Name, *expectation.Type, variable.Type)
	}
	if expectation.HasDefault != nil {
		hasDefault := variable.Default != nil
		if hasDefault != *expectation.HasDefault {
			t.Errorf("Variable %s: expected hasDefault=%t, got %t", variable.Name, *expectation.HasDefault, hasDefault)
		}
	}
	if expectation.Sensitive != nil && variable.Sensitive != *expectation.Sensitive {
		t.Errorf("Variable %s: expected sensitive=%t, got %t", variable.Name, *expectation.Sensitive, variable.Sensitive)
	}
	if expectation.Required != nil && variable.Required != *expectation.Required {
		t.Errorf("Variable %s: expected required=%t, got %t", variable.Name, *expectation.Required, variable.Required)
	}
	if expectation.ValidationCount != nil && len(variable.Validation) != *expectation.ValidationCount {
		t.Errorf("Variable %s: expected %d validation rules, got %d", variable.Name, *expectation.ValidationCount, len(variable.Validation))
	}
}

func validateOutputExpectation(t *testing.T, output *schema.Output, expectation *OutputExpectation) {
	t.Helper()
	if expectation.Sensitive != nil && output.Sensitive != *expectation.Sensitive {
		t.Errorf("Output %s: expected sensitive=%t, got %t", output.Name, *expectation.Sensitive, output.Sensitive)
	}
}

func validateTerraformExpectation(t *testing.T, config *TerraformConfig, expectation *TerraformExpectation) {
	t.Helper()
	if len(config.Terraform) == 0 {
		t.Error("No terraform blocks found")
		return
	}

	terraform := config.Terraform[0]

	if expectation.RequiredVersion != nil && terraform.RequiredVersion != *expectation.RequiredVersion {
		t.Errorf("Expected required_version %s, got %s", *expectation.RequiredVersion, terraform.RequiredVersion)
	}

	if expectation.ProviderCount != nil && len(terraform.RequiredProviders) != *expectation.ProviderCount {
		t.Errorf("Expected %d providers, got %d", *expectation.ProviderCount, len(terraform.RequiredProviders))
	}

	if expectation.ExperimentCount != nil && len(terraform.Experiments) != *expectation.ExperimentCount {
		t.Errorf("Expected %d experiments, got %d", *expectation.ExperimentCount, len(terraform.Experiments))
	}

	// Validate specific providers
	for name, providerExpectation := range expectation.Providers {
		if provider, exists := terraform.RequiredProviders[name]; exists {
			if providerExpectation.Source != nil && provider.Source != *providerExpectation.Source {
				t.Errorf("Provider %s: expected source %s, got %s", name, *providerExpectation.Source, provider.Source)
			}
			if providerExpectation.Version != nil && provider.Version != *providerExpectation.Version {
				t.Errorf("Provider %s: expected version %s, got %s", name, *providerExpectation.Version, provider.Version)
			}
		} else {
			t.Errorf("Provider %s not found", name)
		}
	}
}

// Helper functions to create pointers for expectations
func ptr[T any](v T) *T {
	return &v
}

// New expectation-based tests
func TestVariableBlocks(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		expectError  bool
		expectations TestExpectations
	}{
		{
			name: "Variable with default values",
			files: map[string]string{
				"variables.tf": `
variable "default_string" {
  type        = string
  description = "String variable with default value"
  default     = "default_value"
}

variable "default_object" {
  type = object({
    name = string
    port = number
  })
  description = "Object variable with default value"
  default = {
    name = "web-server"
    port = 8080
  }
}`,
			},
			expectations: TestExpectations{
				VariableCount: ptr(2),
				Variables: map[string]*VariableExpectation{
					"default_string": {
						Type:       ptr("string"),
						HasDefault: ptr(true),
					},
					"default_object": {
						HasDefault: ptr(true),
					},
				},
			},
		},
		{
			name: "Variable with validation rules",
			files: map[string]string{
				"variables.tf": `
variable "validations" {
  type = list(string)
  description = "this is a list variable"

  validation {
    condition     = length(var.validations) > 0
    error_message = "validations1"
  }

  validation {
    condition     = length(var.validations) < 99
    error_message = "validations2"
  }
}`,
			},
			expectations: TestExpectations{
				VariableCount: ptr(1),
				Variables: map[string]*VariableExpectation{
					"validations": {
						ValidationCount: ptr(2),
					},
				},
			},
		},
		{
			name: "Variable with optional attributes",
			files: map[string]string{
				"variables.tf": `
variable "sensitive" {
  type        = string
  description = "this is a sensitive variable"
  sensitive   = true
}

variable "nullable" {
  type        = string
  description = "Variable that can be null"
  default     = null
}`,
			},
			expectations: TestExpectations{
				VariableCount: ptr(2),
				Variables: map[string]*VariableExpectation{
					"sensitive": {
						Sensitive: ptr(true),
					},
					"nullable": {
						Required: ptr(false),
					},
				},
			},
		},
		{
			name: "Multiple variable types",
			files: map[string]string{
				"main.tf": `
variable "string" {
  type = string
}

variable "map" {
  type = map(string)
}

variable "list" {
  type = list(string)
}

variable "set" {
  type = set(string)
}

variable "tuple" {
  type = tuple(string, number)
}

variable "any" {
  type = any
}`,
			},
			expectations: TestExpectations{
				VariableCount: ptr(6),
				Variables: map[string]*VariableExpectation{
					"string": {Type: ptr("string")},
					"map":    {Type: ptr("map(string)")},
					"list":   {Type: ptr("list(string)")},
					"set":    {Type: ptr("set(string)")},
					"tuple":  {Type: ptr("tuple(string, number)")},
					"any":    {Type: ptr("any")},
				},
			},
		},
		{
			name: "Empty variable block",
			files: map[string]string{
				"variables.tf": `variable "empty" {}`,
			},
			expectations: TestExpectations{
				VariableCount: ptr(1),
				Variables: map[string]*VariableExpectation{
					"empty": {}, // Just check existence
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFS := newTestFileSystem(tt.files)
			parser := NewParser(testFS, Simple)
			config, err := parser.ParseTerraformWorkspace(".")

			if tt.expectError && err == nil {
				t.Fatal("Expected error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if config != nil {
				validateExpectations(t, config, tt.expectations)
			}
		})
	}
}

func TestOutputBlocks(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		expectations TestExpectations
	}{
		{
			name: "Complex outputs with expressions",
			files: map[string]string{
				"variables.tf": `
variable "string" {
  type = string
  description = "this is a string variable"
}

variable "list" {
  type = list(string)
  description = "this is a list variable"
}

variable "map" {
  type = map(string)
  description = "this is a map variable"
}`,
				"outputs.tf": `
output "computed" {
  description = "Computed value with interpolation"
  value       = "prefix-${var.string}-suffix"
}

output "complex_expression" {
  description = "Complex expression with function call"
  value       = length(var.list) > 0 ? var.list[0] : "default"
}

output "map_access" {
  description = "Accessing map values"
  value       = var.map["key"]
  sensitive   = false
}`,
			},
			expectations: TestExpectations{
				OutputCount: ptr(3),
				Outputs: map[string]*OutputExpectation{
					"computed":           {},
					"complex_expression": {},
					"map_access": {
						Sensitive: ptr(false),
					},
				},
			},
		},
		{
			name: "Sensitive outputs",
			files: map[string]string{
				"variables.tf": `
variable "secret" {
  type = string
  sensitive = true
}`,
				"outputs.tf": `
output "sensitive_output" {
  description = "This is a sensitive output"
  value       = var.secret
  sensitive   = true
}

output "non_sensitive" {
  description = "This is not sensitive"
  value       = "public_value"
  sensitive   = false
}`,
			},
			expectations: TestExpectations{
				OutputCount: ptr(2),
				Outputs: map[string]*OutputExpectation{
					"sensitive_output": {
						Sensitive: ptr(true),
					},
					"non_sensitive": {
						Sensitive: ptr(false),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFS := newTestFileSystem(tt.files)
			parser := NewParser(testFS, Simple)
			config, err := parser.ParseTerraformWorkspace(".")

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateExpectations(t, config, tt.expectations)
		})
	}
}

func TestTerraformBlocks(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		expectations TestExpectations
	}{
		{
			name: "Terraform block with required providers",
			files: map[string]string{
				"terraform.tf": `
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }

  experiments = ["module_variable_optional_attrs", "config_driven_move"]
}`,
			},
			expectations: TestExpectations{
				TerraformCount: ptr(1),
				TerraformSettings: &TerraformExpectation{
					RequiredVersion: ptr(">= 1.0.0"),
					ProviderCount:   ptr(2),
					ExperimentCount: ptr(2),
					Providers: map[string]*ProviderExpectation{
						"aws": {
							Source:  ptr("hashicorp/aws"),
							Version: ptr("~> 5.0"),
						},
						"azurerm": {
							Source:  ptr("hashicorp/azurerm"),
							Version: ptr("~> 3.0"),
						},
					},
				},
			},
		},
		{
			name: "Minimal terraform block",
			files: map[string]string{
				"terraform.tf": `
terraform {
  required_version = ">= 1.0.0"
}`,
			},
			expectations: TestExpectations{
				TerraformCount: ptr(1),
				TerraformSettings: &TerraformExpectation{
					RequiredVersion: ptr(">= 1.0.0"),
					ProviderCount:   ptr(0),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFS := newTestFileSystem(tt.files)
			parser := NewParser(testFS, Simple)
			config, err := parser.ParseTerraformWorkspace(".")

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateExpectations(t, config, tt.expectations)
		})
	}
}

func TestMixedBlocks(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		expectations TestExpectations
	}{
		{
			name: "Complete terraform configuration",
			files: map[string]string{
				"variables.tf": `
variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  type = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}`,
				"outputs.tf": `
output "region_output" {
  description = "Selected AWS region"
  value       = var.region
}

output "env_output" {
  description = "Environment name"
  value       = var.environment
  sensitive   = false
}`,
				"terraform.tf": `
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}`,
			},
			expectations: TestExpectations{
				VariableCount:  ptr(2),
				OutputCount:    ptr(2),
				TerraformCount: ptr(1),
				Variables: map[string]*VariableExpectation{
					"region": {
						HasDefault: ptr(true),
					},
					"environment": {
						ValidationCount: ptr(1),
					},
				},
				Outputs: map[string]*OutputExpectation{
					"region_output": {},
					"env_output": {
						Sensitive: ptr(false),
					},
				},
				TerraformSettings: &TerraformExpectation{
					RequiredVersion: ptr(">= 1.0"),
					ProviderCount:   ptr(1),
					Providers: map[string]*ProviderExpectation{
						"aws": {
							Version: ptr("~> 5.0"),
						},
					},
				},
			},
		},
		{
			name: "Empty configuration files",
			files: map[string]string{
				"empty.tf": "",
			},
			expectations: TestExpectations{
				VariableCount:  ptr(0),
				OutputCount:    ptr(0),
				TerraformCount: ptr(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFS := newTestFileSystem(tt.files)
			parser := NewParser(testFS, Simple)
			config, err := parser.ParseTerraformWorkspace(".")

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateExpectations(t, config, tt.expectations)
		})
	}
}

func TestParsingLevels(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		mode         Mode
		expectations TestExpectations
	}{
		{
			name: "Simple level - only basic blocks",
			files: map[string]string{
				"main.tf": `
variable "test_var" {
  type = string
}

output "test_output" {
  value = var.test_var
}

terraform {
  required_version = ">= 1.0"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
}`,
			},
			mode: Simple,
			expectations: TestExpectations{
				VariableCount:  ptr(1),
				OutputCount:    ptr(1),
				TerraformCount: ptr(1),
			},
		},
		{
			name: "Detail level - all blocks (when implemented)",
			files: map[string]string{
				"main.tf": `
variable "test_var" {
  type = string
}

output "test_output" {
  value = var.test_var
}

terraform {
  required_version = ">= 1.0"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
}`,
			},
			mode: Detail,
			expectations: TestExpectations{
				VariableCount:  ptr(1),
				OutputCount:    ptr(1),
				TerraformCount: ptr(1),
				// Note: Resource, data, module parsing not implemented yet
			},
		},
		{
			name: "Simple level - mixed configuration",
			files: map[string]string{
				"variables.tf": `
variable "environment" {
  type = string
  validation {
    condition = contains(["dev", "prod"], var.environment)
    error_message = "Must be dev or prod"
  }
}`,
				"outputs.tf": `
output "env" {
  value = var.environment
}`,
				"resources.tf": `
resource "aws_s3_bucket" "example" {
  bucket = "my-bucket"
}

data "aws_caller_identity" "current" {}

module "database" {
  source = "./modules/database"
}`,
				"terraform.tf": `
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}`,
			},
			mode: Simple,
			expectations: TestExpectations{
				VariableCount:  ptr(1),
				OutputCount:    ptr(1),
				TerraformCount: ptr(1),
				Variables: map[string]*VariableExpectation{
					"environment": {
						ValidationCount: ptr(1),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFS := newTestFileSystem(tt.files)
			parser := NewParser(testFS, tt.mode)
			config, err := parser.ParseTerraformWorkspace(".")

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateExpectations(t, config, tt.expectations)
		})
	}
}

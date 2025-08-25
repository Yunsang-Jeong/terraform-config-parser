output "string" {
  description = "This outputs the string variable value"
  value       = var.string
}

output "list" {
  description = "This outputs the list variable"
  value       = var.list
}

output "sensitive" {
  description = "This is a sensitive output"
  value       = var.any
  sensitive   = true
}

output "computed" {
  description = "Computed value with interpolation"
  value       = "prefix-${var.string}-suffix"
}

output "literal_string" {
  description = "A literal string value"
  value       = "hello world"
}

output "literal_number" {
  description = "A literal number value"
  value       = 42
}

output "literal_bool" {
  description = "A literal boolean value"
  value       = false
}

output "complex_expression" {
  description = "Complex expression with function call"
  value       = length(var.list) > 0 ? var.list[0] : "default"
}

output "object_access" {
  description = "Accessing object attributes"
  value       = var.object.name
}

output "map_access" {
  description = "Accessing map values"
  value       = var.map["key"]
  sensitive   = false
}

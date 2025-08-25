variable "empty" {}

variable "string" {
  type = string
  description = "this is a string variable"
}

variable "map" {
  type = map(string)
  description = "this is a map variable"
}

variable "object" {
  type = object({
    name = string
    age  = optional(number)
    tags = map(string)
  })
  description = "this is a object variable"
}

variable "list" {
  type = list(string)
  description = "this is a list variable"
}

variable "set" {
  type = set(string)
  description = "this is a set variable"
}

variable "tuple" {
  type = tuple(string, number)
  description = "this is a tuple variable"
}

variable "any" {
  type = any
  description = "this is an any variable"
}

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
}

variable "sensitive" {
  type        = string
  description = "this is a sensitive variable"
  sensitive   = true
}

variable "default_string" {
  type        = string
  description = "String variable with default value"
  default     = "default_value"
}

variable "default_number" {
  type        = number
  description = "Number variable with default value"
  default     = 100
}

variable "default_bool" {
  type        = bool
  description = "Boolean variable with default value"
  default     = true
}

variable "default_list" {
  type        = list(string)
  description = "List variable with default value"
  default     = ["item1", "item2", "item3"]
}

variable "default_map" {
  type        = map(string)
  description = "Map variable with default value"
  default     = {
    env = "development"
    app = "myapp"
  }
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
}

variable "nullable" {
  type        = string
  description = "Variable that can be null"
  default     = null
}
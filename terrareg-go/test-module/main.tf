variable "name" {
  description = "The name of the resource"
  type        = string
}

variable "enabled" {
  description = "Whether to enable the resource"
  type        = bool
  default     = true
}

resource "null_resource" "example" {
  count = var.enabled ? 1 : 0
  
  triggers = {
    name = var.name
  }
}

output "id" {
  description = "The ID of the created resource"
  value       = try(null_resource.example[0].id, null)
}

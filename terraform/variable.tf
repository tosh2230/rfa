variable "project" {
  description = "A name of a GCP Project"
  type        = string
  default     = null
}

variable "region" {
  description = "A region to use the module"
  type        = string
  default     = "us-east1"

  # validation {
  #   condition     = var.region == "us-east1"
  #   error_message = "The region must be us-east1."
  # }
}

variable "zone" {
  description = "A zone to use the module"
  type        = string
  default     = "us-east1-a"

  # validation {
  #   condition     = contains(["us-east1-a", "us-east1-b", "us-east1-c"], var.zone)
  #   error_message = "The zone must be in us-east1 region."
  # }
}

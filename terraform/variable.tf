variable "project" {
  description = "A name of a GCP Project"
  type        = string
  default     = null
}

variable "region" {
  description = "A region to use the module"
  type        = string
  default     = "asia-northeast1"

  # validation {
  #   condition     = var.region == "asia-northeast1"
  #   error_message = "The region must be asia-northeast1."
  # }
}

variable "zone" {
  description = "A zone to use the module"
  type        = string
  default     = "asia-northeast1-a"

  # validation {
  #   condition     = contains(["asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"], var.zone)
  #   error_message = "The zone must be in asia-northeast1 region."
  # }
}

variable "zip-bucket" {
  description = "A GCS bucket for sourcecodes of cloud functions"
  type        = string
  default     = "${var.project}-zip-bucket"
}

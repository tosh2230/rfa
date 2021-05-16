provider "google" {
  version = "3.60.0"
  project = var.project
  region  = var.region
  zone    = var.zone
}

terraform {
  backend "gcs" {
  }
}

##############################################
# Service Account
##############################################
resource "google_service_account" "sa_functions_rfa" {
  account_id   = "sa-functions-rfa"
  display_name = "sa-functions-rfa"
}

##############################################
# Cloud Functions
# https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function
##############################################
resource "google_cloudfunctions_function" "rfa" {
  name                  = "rfa"
  description           = "https://github.com/tosh223/rfa"
  runtime               = "go113"
  source_archive_bucket = var.zip_bucket
  source_archive_object = google_storage_bucket_object.function_rfa_packages.name
  available_memory_mb   = 256
  timeout               = 30
  entry_point           = "EntryPointHTTP"
  trigger_http          = true
  service_account_email = google_service_account.sa_functions_rfa.email
}

resource "google_storage_bucket_object" "function_rfa_packages" {
  name   = "packages/go/function_rfa.${data.archive_file.function_rfa_archive.output_md5}.zip"
  bucket = var.zip_bucket
  source = data.archive_file.function_rfa_archive.output_path
}

data "archive_file" "function_rfa_archive" {
  type        = "zip"
  source_dir  = ".."
  output_path = "zip/go/rfa.zip"
}

resource "google_cloudfunctions_function_iam_member" "rfa_member" {
  project        = google_cloudfunctions_function.rfa.project
  region         = google_cloudfunctions_function.rfa.region
  cloud_function = google_cloudfunctions_function.rfa.name
  role           = "roles/cloudfunctions.invoker"
  member         = "serviceAccount:${google_service_account.sa_functions_rfa.email}"
}

##############################################
# Output
##############################################
output "function_rfa_url" {
  value = google_cloudfunctions_function.rfa.https_trigger_url
}

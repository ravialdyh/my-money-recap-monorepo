terraform {
  required_version = ">= 1.12.0"

  required_providers {
    keycloak = {
      source  = "keycloak/keycloak"
      version = "5.9.0"
    }
  }
}

provider "keycloak" {}

variable "ssl_required" {
  description = "Keycloak's sslRequired for the mmr realm: none on the laptop, where every caller speaks plain HTTP, external on the machine. The provider defaults it to external, so this root never lets it default (ARCHITECTURE A13)."
  type        = string
}

variable "bff_client_secret" {
  description = "The bff client's secret. Explicit rather than generated, because the BFF signs in as this client for the Login Token endpoint (BFF_CLIENT_SECRET), and so do a10-smoke.sh and the members wrong-audience test (A10)."
  type        = string
  sensitive   = true
}

variable "notify_client_secret" {
  description = "The notify client's secret: notify's schedules fetch its service-account token, whose audience is the money service (ARCHITECTURE §4.2), with the client-credentials grant (NOTIFY_CLIENT_SECRET)."
  type        = string
  sensitive   = true
}

variable "members_client_secret" {
  description = "The members client's secret: the deletion job deletes Keycloak accounts with its service account, which holds manage-users (ARCHITECTURE §7.1, MEMBERS_CLIENT_SECRET)."
  type        = string
  sensitive   = true
}

variable "test_user_password" {
  description = "The realm test Member's password, the seam the Connect-boundary suite and the iOS UI test sign in with. Empty creates no test Member, which is the machine; a realm with ssl_required none, the laptop, refuses empty."
  type        = string
  sensitive   = true

  validation {
    condition     = var.test_user_password != "" || var.ssl_required != "none"
    error_message = "TEST_USER_PASSWORD is empty: set it in infra/compose/.env. Only the machine, whose realm has ssl_required external, goes without the test Member."
  }
}

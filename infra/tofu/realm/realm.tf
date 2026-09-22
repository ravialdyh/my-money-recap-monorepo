resource "keycloak_realm" "mmr" {
  realm                       = "mmr"
  enabled                     = true
  ssl_required                = var.ssl_required
  registration_allowed        = false
  login_with_email_allowed    = true
  default_signature_algorithm = "RS256"

  offline_session_idle_timeout         = "720h0m0s"
  offline_session_max_lifespan_enabled = false
}

resource "keycloak_openid_client_scope" "login_token" {
  realm_id               = keycloak_realm.mmr.id
  name                   = "login-token"
  description            = "Requested by the app when it redeems a Login Token; routes the direct grant to the Login Token branch (A10)."
  include_in_token_scope = false
}

resource "keycloak_openid_client" "ios" {
  realm_id                     = keycloak_realm.mmr.id
  client_id                    = "ios"
  name                         = "My Money Recap iOS"
  enabled                      = true
  access_type                  = "PUBLIC"
  standard_flow_enabled        = true
  direct_access_grants_enabled = true
  implicit_flow_enabled        = false
  pkce_code_challenge_method   = "S256"
  valid_redirect_uris          = ["mymoneyrecap://auth/callback"]
  web_origins                  = []

  authentication_flow_binding_overrides {
    direct_grant_id = keycloak_authentication_flow.ios_direct_grant.id
  }
}

resource "keycloak_openid_audience_protocol_mapper" "bff_audience" {
  realm_id                   = keycloak_realm.mmr.id
  client_id                  = keycloak_openid_client.ios.id
  name                       = "bff-audience"
  included_client_audience   = keycloak_openid_client.bff.client_id
  add_to_id_token            = false
  add_to_access_token        = true
  add_to_token_introspection = true
}

resource "keycloak_openid_client_default_scopes" "ios" {
  realm_id       = keycloak_realm.mmr.id
  client_id      = keycloak_openid_client.ios.id
  default_scopes = ["profile", "email", "roles", "web-origins", "acr", "basic"]
}

resource "keycloak_openid_client_optional_scopes" "ios" {
  realm_id  = keycloak_realm.mmr.id
  client_id = keycloak_openid_client.ios.id
  optional_scopes = [
    keycloak_openid_client_scope.login_token.name,
    "address",
    "phone",
    "offline_access",
    "organization",
    "microprofile-jwt",
  ]
}

resource "keycloak_openid_client" "bff" {
  realm_id                     = keycloak_realm.mmr.id
  client_id                    = "bff"
  name                         = "My Money Recap BFF"
  enabled                      = true
  access_type                  = "CONFIDENTIAL"
  client_authenticator_type    = "client-secret"
  client_secret                = var.bff_client_secret
  standard_flow_enabled        = false
  direct_access_grants_enabled = false
  implicit_flow_enabled        = false
  service_accounts_enabled     = true
  valid_redirect_uris          = []
  web_origins                  = []
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = keycloak_realm.mmr.id
  client_id = "realm-management"
}

resource "keycloak_openid_client_service_account_role" "bff_manage_users" {
  realm_id                = keycloak_realm.mmr.id
  service_account_user_id = keycloak_openid_client.bff.service_account_user_id
  client_id               = data.keycloak_openid_client.realm_management.id
  role                    = "manage-users"
}

resource "keycloak_user" "test" {
  count          = var.test_user_password == "" ? 0 : 1
  realm_id       = keycloak_realm.mmr.id
  username       = "test"
  email          = "test@example.com"
  first_name     = "Test"
  last_name      = "Member"
  enabled        = true
  email_verified = true

  initial_password {
    value     = var.test_user_password
    temporary = false
  }
}

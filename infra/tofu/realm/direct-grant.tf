resource "keycloak_authentication_flow" "ios_direct_grant" {
  realm_id    = keycloak_realm.mmr.id
  alias       = "ios direct grant"
  provider_id = "basic-flow"
  description = "Direct grant for the ios client: the Login Token branch when scope contains login-token, the username-password branch otherwise. The Login Token authenticator fails the flow when login_token is absent, so the two cannot be plain alternatives (A10)."
}

resource "keycloak_authentication_subflow" "login_token" {
  realm_id          = keycloak_realm.mmr.id
  alias             = "ios direct grant - login token"
  parent_flow_alias = keycloak_authentication_flow.ios_direct_grant.alias
  provider_id       = "basic-flow"
  requirement       = "CONDITIONAL"
  priority          = 10
}

resource "keycloak_authentication_execution" "login_token_scope" {
  realm_id          = keycloak_realm.mmr.id
  parent_flow_alias = keycloak_authentication_subflow.login_token.alias
  authenticator     = "conditional-client-scope"
  requirement       = "REQUIRED"
  priority          = 10
}

resource "keycloak_authentication_execution_config" "login_token_requested" {
  realm_id     = keycloak_realm.mmr.id
  execution_id = keycloak_authentication_execution.login_token_scope.id
  alias        = "login-token requested"

  config = {
    client_scope = keycloak_openid_client_scope.login_token.name
    negate       = "false"
  }
}

resource "keycloak_authentication_execution" "login_token_grant" {
  realm_id          = keycloak_realm.mmr.id
  parent_flow_alias = keycloak_authentication_subflow.login_token.alias
  authenticator     = "direct-grant-login-token"
  requirement       = "REQUIRED"
  priority          = 20

  depends_on = [keycloak_authentication_execution.login_token_scope]
}

resource "keycloak_authentication_subflow" "password" {
  realm_id          = keycloak_realm.mmr.id
  alias             = "ios direct grant - password"
  parent_flow_alias = keycloak_authentication_flow.ios_direct_grant.alias
  provider_id       = "basic-flow"
  requirement       = "CONDITIONAL"
  priority          = 20

  depends_on = [keycloak_authentication_subflow.login_token]
}

resource "keycloak_authentication_execution" "password_scope" {
  realm_id          = keycloak_realm.mmr.id
  parent_flow_alias = keycloak_authentication_subflow.password.alias
  authenticator     = "conditional-client-scope"
  requirement       = "REQUIRED"
  priority          = 10
}

resource "keycloak_authentication_execution_config" "login_token_not_requested" {
  realm_id     = keycloak_realm.mmr.id
  execution_id = keycloak_authentication_execution.password_scope.id
  alias        = "login-token not requested"

  config = {
    client_scope = keycloak_openid_client_scope.login_token.name
    negate       = "true"
  }
}

resource "keycloak_authentication_execution" "password_username" {
  realm_id          = keycloak_realm.mmr.id
  parent_flow_alias = keycloak_authentication_subflow.password.alias
  authenticator     = "direct-grant-validate-username"
  requirement       = "REQUIRED"
  priority          = 20

  depends_on = [keycloak_authentication_execution.password_scope]
}

resource "keycloak_authentication_execution" "password_password" {
  realm_id          = keycloak_realm.mmr.id
  parent_flow_alias = keycloak_authentication_subflow.password.alias
  authenticator     = "direct-grant-validate-password"
  requirement       = "REQUIRED"
  priority          = 30

  depends_on = [keycloak_authentication_execution.password_username]
}

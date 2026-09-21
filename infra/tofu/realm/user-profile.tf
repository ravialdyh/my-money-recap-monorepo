resource "keycloak_realm_user_profile" "mmr" {
  realm_id = keycloak_realm.mmr.id

  attribute {
    name         = "username"
    display_name = "$${username}"

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name   = "length"
      config = { min = "3", max = "255" }
    }

    validator {
      name = "username-prohibited-characters"
    }

    validator {
      name = "up-username-not-idn-homograph"
    }
  }

  attribute {
    name               = "email"
    display_name       = "$${email}"
    required_for_roles = ["user"]

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name = "email"
    }

    validator {
      name   = "length"
      config = { max = "255" }
    }
  }

  attribute {
    name         = "firstName"
    display_name = "$${firstName}"

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name   = "length"
      config = { max = "255" }
    }

    validator {
      name = "person-name-prohibited-characters"
    }
  }

  attribute {
    name         = "lastName"
    display_name = "$${lastName}"

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name   = "length"
      config = { max = "255" }
    }

    validator {
      name = "person-name-prohibited-characters"
    }
  }

  attribute {
    name         = "org.ro.active"
    display_name = "Active organization ID"

    permissions {
      view = ["admin"]
      edit = ["admin"]
    }
  }

  group {
    name                = "user-metadata"
    display_header      = "User metadata"
    display_description = "Attributes, which refer to user metadata"
  }
}

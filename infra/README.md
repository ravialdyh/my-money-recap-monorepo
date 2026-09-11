# infra

Deployment and infrastructure-as-code. Nothing here yet.

Planned, per `docs/ARCHITECTURE.md` §10 and §7.5: the phase-1 `compose.yaml` (Nginx, certbot,
the BFF, the four services, Keycloak, Postgres, RabbitMQ, Redis), the Nginx configuration, and
the Keycloak realm export at `infra/keycloak/realm.json`. Secrets never live here; they come from
the environment.

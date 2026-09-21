ALTER TABLE magic_link_send DROP CONSTRAINT magic_link_send_pkey;

ALTER TABLE magic_link_send ADD PRIMARY KEY (email, idempotency_key);

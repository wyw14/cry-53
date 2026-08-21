BEGIN;
CREATE TABLE IF NOT EXISTS config_types (name text PRIMARY KEY, definition jsonb NOT NULL, updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS configurations (id text PRIMARY KEY, config_key text NOT NULL UNIQUE, document jsonb NOT NULL, revision bigint NOT NULL CHECK (revision > 0), updated_at timestamptz NOT NULL);
CREATE TABLE IF NOT EXISTS bundles (id text PRIMARY KEY, document jsonb NOT NULL, updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS configuration_versions (configuration_id text NOT NULL REFERENCES configurations(id), version_number integer NOT NULL, document jsonb NOT NULL, PRIMARY KEY (configuration_id, version_number));
CREATE TABLE IF NOT EXISTS audit_events (id text PRIMARY KEY, request_id text NOT NULL, actor_id text NOT NULL, action text NOT NULL, resource text NOT NULL, resource_id text NOT NULL, document jsonb NOT NULL, created_at timestamptz NOT NULL);
CREATE INDEX IF NOT EXISTS audit_events_resource_idx ON audit_events(resource, resource_id, created_at DESC);
CREATE TABLE IF NOT EXISTS configuration_usages (id text PRIMARY KEY, configuration_id text NOT NULL REFERENCES configurations(id), document jsonb NOT NULL, last_observed_at timestamptz NOT NULL);
CREATE TABLE IF NOT EXISTS idempotency_records (key text PRIMARY KEY, request_hash text NOT NULL, resource_id text NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
COMMIT;


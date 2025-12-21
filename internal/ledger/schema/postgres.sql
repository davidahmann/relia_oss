-- =========================
-- Enums
-- =========================
DO $$ BEGIN
  CREATE TYPE relia_idem_status AS ENUM
    ('pending_approval','approved_ready','issuing','allowed','denied','errored');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE relia_approval_status AS ENUM
    ('pending','approved','denied');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE relia_outcome_status AS ENUM
    ('approval_pending','approval_approved','approval_denied',
     'issuing_credentials','issued_credentials','denied','issue_failed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- =========================
-- Keys
-- =========================
CREATE TABLE IF NOT EXISTS relia_keys (
  key_id      TEXT PRIMARY KEY,
  public_key  BYTEA NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL,
  rotated_at  TIMESTAMPTZ
);

-- =========================
-- Policy versions
-- =========================
CREATE TABLE IF NOT EXISTS relia_policy_versions (
  policy_hash    TEXT PRIMARY KEY,
  policy_id      TEXT NOT NULL,
  policy_version TEXT NOT NULL,
  policy_yaml    TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL
);

-- =========================
-- Contexts + Decisions
-- =========================
CREATE TABLE IF NOT EXISTS relia_contexts (
  context_id  TEXT PRIMARY KEY,
  created_at  TIMESTAMPTZ NOT NULL,
  body_json   JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS relia_decisions (
  decision_id TEXT PRIMARY KEY,
  created_at  TIMESTAMPTZ NOT NULL,
  context_id  TEXT NOT NULL REFERENCES relia_contexts(context_id),
  policy_hash TEXT NOT NULL REFERENCES relia_policy_versions(policy_hash),
  verdict     TEXT NOT NULL CHECK (verdict IN ('allow','deny','require_approval')),
  body_json   JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rel_decisions_context ON relia_decisions(context_id);
CREATE INDEX IF NOT EXISTS idx_rel_decisions_policy  ON relia_decisions(policy_hash);

-- =========================
-- Idempotency keys
-- =========================
CREATE TABLE IF NOT EXISTS relia_idempotency_keys (
  idem_key          TEXT PRIMARY KEY,
  status            relia_idem_status NOT NULL,
  approval_id       TEXT,
  latest_receipt_id TEXT,
  final_receipt_id  TEXT,
  created_at        TIMESTAMPTZ NOT NULL,
  updated_at        TIMESTAMPTZ NOT NULL,
  ttl_expires_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_rel_idem_status ON relia_idempotency_keys(status);

-- =========================
-- Approvals
-- =========================
CREATE TABLE IF NOT EXISTS relia_approvals (
  approval_id    TEXT PRIMARY KEY,
  idem_key       TEXT NOT NULL UNIQUE REFERENCES relia_idempotency_keys(idem_key),
  status         relia_approval_status NOT NULL,
  slack_channel  TEXT,
  slack_msg_ts   TEXT,
  approved_by    TEXT,
  approved_at    TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL,
  updated_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rel_approvals_status ON relia_approvals(status);

ALTER TABLE relia_idempotency_keys
  ADD CONSTRAINT fk_idem_approval
  FOREIGN KEY (approval_id) REFERENCES relia_approvals(approval_id);

-- =========================
-- Receipts
-- =========================
CREATE TABLE IF NOT EXISTS relia_receipts (
  receipt_id             TEXT PRIMARY KEY,
  idem_key               TEXT NOT NULL REFERENCES relia_idempotency_keys(idem_key),
  created_at             TIMESTAMPTZ NOT NULL,

  supersedes_receipt_id  TEXT REFERENCES relia_receipts(receipt_id),

  context_id             TEXT NOT NULL REFERENCES relia_contexts(context_id),
  decision_id            TEXT NOT NULL REFERENCES relia_decisions(decision_id),
  policy_hash            TEXT NOT NULL REFERENCES relia_policy_versions(policy_hash),

  approval_id            TEXT REFERENCES relia_approvals(approval_id),

  outcome_status         relia_outcome_status NOT NULL,
  final                  BOOLEAN NOT NULL,
  expires_at             TIMESTAMPTZ,

  body_json              JSONB NOT NULL,
  body_digest            TEXT NOT NULL,
  key_id                 TEXT NOT NULL REFERENCES relia_keys(key_id),
  sig                    BYTEA NOT NULL,

  CONSTRAINT chk_body_digest_matches CHECK (body_digest = receipt_id)
);

CREATE INDEX IF NOT EXISTS idx_rel_receipts_idem_created ON relia_receipts(idem_key, created_at);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_supersedes   ON relia_receipts(supersedes_receipt_id);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_outcome      ON relia_receipts(outcome_status);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_context      ON relia_receipts(context_id);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_decision     ON relia_receipts(decision_id);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_policy       ON relia_receipts(policy_hash);
CREATE INDEX IF NOT EXISTS idx_rel_receipts_final        ON relia_receipts(final);

ALTER TABLE relia_idempotency_keys
  ADD CONSTRAINT fk_idem_latest_receipt
  FOREIGN KEY (latest_receipt_id) REFERENCES relia_receipts(receipt_id);

ALTER TABLE relia_idempotency_keys
  ADD CONSTRAINT fk_idem_final_receipt
  FOREIGN KEY (final_receipt_id) REFERENCES relia_receipts(receipt_id);

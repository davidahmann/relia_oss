PRAGMA foreign_keys = ON;

-- =========================
-- Keys
-- =========================
CREATE TABLE IF NOT EXISTS keys (
  key_id      TEXT PRIMARY KEY,
  public_key  BLOB NOT NULL,
  created_at  TEXT NOT NULL,
  rotated_at  TEXT
);

-- =========================
-- Policy versions
-- =========================
CREATE TABLE IF NOT EXISTS policy_versions (
  policy_hash    TEXT PRIMARY KEY,
  policy_id      TEXT NOT NULL,
  policy_version TEXT NOT NULL,
  policy_yaml    TEXT NOT NULL,
  created_at     TEXT NOT NULL
);

-- =========================
-- Contexts + Decisions
-- =========================
CREATE TABLE IF NOT EXISTS contexts (
  context_id  TEXT PRIMARY KEY,
  created_at  TEXT NOT NULL,
  body_json   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS decisions (
  decision_id TEXT PRIMARY KEY,
  created_at  TEXT NOT NULL,
  context_id  TEXT NOT NULL,
  policy_hash TEXT NOT NULL,
  verdict     TEXT NOT NULL, -- allow | deny | require_approval
  body_json   TEXT NOT NULL,
  FOREIGN KEY(context_id) REFERENCES contexts(context_id),
  FOREIGN KEY(policy_hash) REFERENCES policy_versions(policy_hash)
);

CREATE INDEX IF NOT EXISTS idx_decisions_context ON decisions(context_id);
CREATE INDEX IF NOT EXISTS idx_decisions_policy  ON decisions(policy_hash);

-- =========================
-- Idempotency keys
-- =========================
CREATE TABLE IF NOT EXISTS idempotency_keys (
  idem_key          TEXT PRIMARY KEY,
  status            TEXT NOT NULL CHECK (status IN (
                      'pending_approval','approved_ready','issuing','allowed','denied','errored'
                    )),
  approval_id       TEXT,
  latest_receipt_id TEXT,
  final_receipt_id  TEXT,
  created_at        TEXT NOT NULL,
  updated_at        TEXT NOT NULL,
  ttl_expires_at    TEXT,

  FOREIGN KEY(approval_id) REFERENCES approvals(approval_id),
  FOREIGN KEY(latest_receipt_id) REFERENCES receipts(receipt_id),
  FOREIGN KEY(final_receipt_id)  REFERENCES receipts(receipt_id)
);

CREATE INDEX IF NOT EXISTS idx_idem_status ON idempotency_keys(status);

-- =========================
-- Approvals
-- =========================
CREATE TABLE IF NOT EXISTS approvals (
  approval_id    TEXT PRIMARY KEY,
  idem_key       TEXT NOT NULL UNIQUE,
  status         TEXT NOT NULL CHECK (status IN ('pending','approved','denied')),
  slack_channel  TEXT,
  slack_msg_ts   TEXT,

  approved_by    TEXT,
  approved_at    TEXT,

  created_at     TEXT NOT NULL,
  updated_at     TEXT NOT NULL,

  FOREIGN KEY(idem_key) REFERENCES idempotency_keys(idem_key)
);

CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status);

-- =========================
-- Receipts
-- =========================
CREATE TABLE IF NOT EXISTS receipts (
  receipt_id             TEXT PRIMARY KEY,
  idem_key               TEXT NOT NULL,
  created_at             TEXT NOT NULL,

  supersedes_receipt_id  TEXT,

  context_id             TEXT NOT NULL,
  decision_id            TEXT NOT NULL,
  policy_hash            TEXT NOT NULL,

  approval_id            TEXT,
  outcome_status         TEXT NOT NULL CHECK (outcome_status IN (
                         'approval_pending','approval_approved','approval_denied',
                         'issuing_credentials','issued_credentials',
                         'denied','issue_failed'
                       )),
  final                  INTEGER NOT NULL CHECK (final IN (0,1)),
  expires_at             TEXT,

  body_json              TEXT NOT NULL,
  body_digest            TEXT NOT NULL,
  key_id                 TEXT NOT NULL,
  sig                    BLOB NOT NULL,

  FOREIGN KEY(idem_key) REFERENCES idempotency_keys(idem_key),
  FOREIGN KEY(supersedes_receipt_id) REFERENCES receipts(receipt_id),
  FOREIGN KEY(context_id) REFERENCES contexts(context_id),
  FOREIGN KEY(decision_id) REFERENCES decisions(decision_id),
  FOREIGN KEY(policy_hash) REFERENCES policy_versions(policy_hash),
  FOREIGN KEY(approval_id) REFERENCES approvals(approval_id),
  FOREIGN KEY(key_id) REFERENCES keys(key_id)
);

CREATE INDEX IF NOT EXISTS idx_receipts_idem_created ON receipts(idem_key, created_at);
CREATE INDEX IF NOT EXISTS idx_receipts_supersedes   ON receipts(supersedes_receipt_id);
CREATE INDEX IF NOT EXISTS idx_receipts_outcome      ON receipts(outcome_status);
CREATE INDEX IF NOT EXISTS idx_receipts_context      ON receipts(context_id);
CREATE INDEX IF NOT EXISTS idx_receipts_decision     ON receipts(decision_id);
CREATE INDEX IF NOT EXISTS idx_receipts_policy       ON receipts(policy_hash);
CREATE INDEX IF NOT EXISTS idx_receipts_final        ON receipts(final);

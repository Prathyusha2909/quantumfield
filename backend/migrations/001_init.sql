CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(120) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'analyst'
);

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(253) NOT NULL,
    port INTEGER NOT NULL DEFAULT 443 CHECK (port BETWEEN 1 AND 65535),
    label VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    last_scanned_at TIMESTAMPTZ,
    current_risk_score INTEGER NOT NULL DEFAULT 0,
    current_pqc_score INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT assets_user_domain_port_key UNIQUE (user_id, domain, port)
);

CREATE TABLE IF NOT EXISTS scans (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    risk_score INTEGER NOT NULL DEFAULT 0,
    pqc_score INTEGER NOT NULL DEFAULT 0,
    tls_version VARCHAR(32) NOT NULL DEFAULT '',
    cipher_suite VARCHAR(128) NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    scan_id UUID NOT NULL UNIQUE REFERENCES scans(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    subject TEXT NOT NULL DEFAULT '',
    common_name VARCHAR(255) NOT NULL DEFAULT '',
    issuer TEXT NOT NULL DEFAULT '',
    serial_number VARCHAR(255) NOT NULL DEFAULT '',
    not_before TIMESTAMPTZ NOT NULL,
    not_after TIMESTAMPTZ NOT NULL,
    public_key_algorithm VARCHAR(64) NOT NULL DEFAULT '',
    signature_algorithm VARCHAR(128) NOT NULL DEFAULT '',
    key_size INTEGER NOT NULL DEFAULT 0,
    chain_valid BOOLEAN NOT NULL DEFAULT FALSE,
    hostname_valid BOOLEAN NOT NULL DEFAULT FALSE,
    self_signed BOOLEAN NOT NULL DEFAULT FALSE,
    subject_alt_names JSONB NOT NULL DEFAULT '[]'::jsonb,
    fingerprint_sha256 VARCHAR(95) NOT NULL DEFAULT '',
    pem TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS findings (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    type VARCHAR(80) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    title VARCHAR(180) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    remediation TEXT NOT NULL DEFAULT '',
    evidence TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'open'
);

CREATE TABLE IF NOT EXISTS pqc_assessments (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    scan_id UUID NOT NULL UNIQUE REFERENCES scans(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    grade VARCHAR(8) NOT NULL DEFAULT '',
    quantum_vulnerable BOOLEAN NOT NULL DEFAULT FALSE,
    rsa_dependency BOOLEAN NOT NULL DEFAULT FALSE,
    ecc_dependency BOOLEAN NOT NULL DEFAULT FALSE,
    tls13_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    legacy_tls_supported BOOLEAN NOT NULL DEFAULT FALSE,
    certificate_rotation_ready BOOLEAN NOT NULL DEFAULT FALSE,
    rationale JSONB NOT NULL DEFAULT '[]'::jsonb
);

-- Dawnix Database Initialization Script
-- PostgreSQL 15+

-- Create schema (if not exists)
CREATE SCHEMA IF NOT EXISTS dawnix;

-- Set search path
SET search_path TO dawnix, public;

-- Process Definition Table
CREATE TABLE IF NOT EXISTS dawnix_process_definition (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    version INT DEFAULT 1,
    structure JSONB NOT NULL,
    form_definition JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT unique_process_code UNIQUE (code)
);

-- Process Instance Table
CREATE TABLE IF NOT EXISTS dawnix_process_instance (
    id BIGSERIAL PRIMARY KEY,
    definition_id BIGINT NOT NULL,
    process_code VARCHAR(255) NOT NULL,
    snapshot_structure JSONB NOT NULL,
    parent_id BIGINT,
    parent_node_id VARCHAR(255),
    form_data JSONB,
    status VARCHAR(50) DEFAULT 'PENDING',
    submitter_id VARCHAR(255),
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_definition_id FOREIGN KEY (definition_id) 
        REFERENCES dawnix_process_definition(id) ON DELETE CASCADE
);

-- Execution Table
CREATE TABLE IF NOT EXISTS dawnix_execution (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL,
    process_code VARCHAR(255),
    node_id VARCHAR(255),
    node_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_instance_id FOREIGN KEY (instance_id) 
        REFERENCES dawnix_process_instance(id) ON DELETE CASCADE
);

-- Process Task Table
CREATE TABLE IF NOT EXISTS dawnix_process_task (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL,
    execution_id BIGINT,
    node_id VARCHAR(255),
    type VARCHAR(50) DEFAULT 'user_task',
    assignee VARCHAR(255),
    candidates VARCHAR(255)[] DEFAULT '{}'::varchar(255)[],
    status VARCHAR(50) DEFAULT 'PENDING',
    action VARCHAR(50),
    comment TEXT,
    form_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_instance_task FOREIGN KEY (instance_id) 
        REFERENCES dawnix_process_instance(id) ON DELETE CASCADE,
    CONSTRAINT fk_execution_task FOREIGN KEY (execution_id) 
        REFERENCES dawnix_execution(id) ON DELETE CASCADE
);

-- User Table (identity is unified as string user_id)
CREATE TABLE IF NOT EXISTS dawnix_users (
    user_id VARCHAR(64) PRIMARY KEY,
    display_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by VARCHAR(64),
    updated_by VARCHAR(64)
);

-- Auth Identity Table (KISS: login/logout, local_password only in current phase)
CREATE TABLE IF NOT EXISTS dawnix_auth_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    provider VARCHAR(64) NOT NULL,
    provider_sub VARCHAR(128) NOT NULL,
    credential_hash VARCHAR(255),
    meta JSONB DEFAULT '{}'::jsonb,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT uk_auth_provider_sub UNIQUE (provider, provider_sub),
    CONSTRAINT fk_auth_user FOREIGN KEY (user_id)
        REFERENCES dawnix_users(user_id) ON DELETE CASCADE
);

-- Create Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_process_definition_code 
    ON dawnix_process_definition(code) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_definition_active 
    ON dawnix_process_definition(is_active) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_instance_definition 
    ON dawnix_process_instance(definition_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_instance_code 
    ON dawnix_process_instance(process_code) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_instance_submitter 
    ON dawnix_process_instance(submitter_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_instance_status 
    ON dawnix_process_instance(status) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_execution_instance 
    ON dawnix_execution(instance_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_execution_node 
    ON dawnix_execution(node_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_task_instance 
    ON dawnix_process_task(instance_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_task_execution 
    ON dawnix_process_task(execution_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_task_assignee 
    ON dawnix_process_task(assignee) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_process_task_status 
    ON dawnix_process_task(status) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_status
    ON dawnix_users(status) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_auth_identity_user
    ON dawnix_auth_identities(user_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_auth_identity_provider
    ON dawnix_auth_identities(provider) WHERE deleted_at IS NULL;

-- Insert sample process definition
INSERT INTO dawnix_process_definition (code, name, version, structure, form_definition, is_active)
VALUES (
    'leave_request',
    '请假审批流程',
    1,
    '{
        "nodes": [
            {"id": "start", "type": "start_event", "name": "开始", "candidates": {}, "properties": null},
            {"id": "manager_review", "type": "user_task", "name": "经理审批", "candidates": {"candidate_users": [], "candidate_groups": []}, "properties": {"assignee_rule": "FIRST_ONE"}},
            {"id": "end", "type": "end_event", "name": "结束", "candidates": {}, "properties": null}
        ],
        "edges": [
            {"id": "edge_1", "source": "start", "target": "manager_review", "condition": "", "is_default": false},
            {"id": "edge_2", "source": "manager_review", "target": "end", "condition": "", "is_default": false}
        ],
        "viewport": {"x": 0, "y": 0, "zoom": 1}
    }'::jsonb,
    '[
        {"key": "days", "type": "number", "value": 0},
        {"key": "reason", "type": "string", "value": ""}
    ]'::jsonb,
    true
) ON CONFLICT (code) DO NOTHING;

-- Insert sample auth user (password: password)
INSERT INTO dawnix_users (user_id, display_name, status, created_by, updated_by)
VALUES ('u_admin', '管理员', 'ACTIVE', 'system', 'system')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO dawnix_auth_identities (user_id, provider, provider_sub, credential_hash)
VALUES (
    'u_admin',
    'local_password',
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
)
ON CONFLICT (provider, provider_sub) DO NOTHING;

-- Log initialization
RAISE NOTICE 'Dawnix database initialized successfully at %', CURRENT_TIMESTAMP;

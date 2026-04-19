-- 000002_create_prompt_management_table.up.sql
CREATE TABLE IF NOT EXISTS prompt_management (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client TEXT NOT NULL,
    use_case TEXT NOT NULL,
    document_type TEXT NOT NULL,
    category TEXT NOT NULL,
    stage_name TEXT NOT NULL,
    active_item_id UUID, -- Will be set/updated by Item Service
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Enforce uniqueness on the business key to prevent duplicates
    UNIQUE(client, use_case, document_type)
);

CREATE INDEX IF NOT EXISTS idx_prompt_management_client ON prompt_management (client) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_prompt_management_use_case ON prompt_management (use_case) WHERE deleted_at IS NULL;

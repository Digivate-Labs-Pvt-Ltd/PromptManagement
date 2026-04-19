CREATE TABLE IF NOT EXISTS prompt_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    management_id UUID NOT NULL REFERENCES prompt_management(id) ON DELETE CASCADE,
    question_key VARCHAR(255) NOT NULL,
    prompt_text TEXT NOT NULL,
    vector_prompt TEXT,
    generation_config JSONB,
    response_schema JSONB,
    top_k NUMERIC,
    ranking_method VARCHAR(50),
    version_no VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    change_log TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT max_len_version CHECK (char_length(version_no) <= 50),
    CONSTRAINT max_len_status CHECK (char_length(status) <= 50),
    UNIQUE (management_id, question_key, version_no)
);

CREATE INDEX IF NOT EXISTS idx_prompt_item_management_id ON prompt_item(management_id);
CREATE INDEX IF NOT EXISTS idx_prompt_item_question_key ON prompt_item(question_key);
CREATE INDEX IF NOT EXISTS idx_prompt_item_status ON prompt_item(status);

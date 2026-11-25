-- Favorites Table
CREATE TABLE IF NOT EXISTS favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_name VARCHAR(50) NOT NULL,
    provider_media_id INT NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    poster_url TEXT,
    rating NUMERIC(3,1) DEFAULT 0.0,
    release_year INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_favorites_user_media UNIQUE (user_id, provider_name, provider_media_id, media_type)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);

-- Custom Lists Table
CREATE TABLE IF NOT EXISTS custom_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_custom_lists_user_id ON custom_lists(user_id);

-- Custom List Items Table (Watch Status: plan_to_watch, watching, watched)
CREATE TABLE IF NOT EXISTS custom_list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES custom_lists(id) ON DELETE CASCADE,
    provider_name VARCHAR(50) NOT NULL,
    provider_media_id INT NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    poster_url TEXT,
    watch_status VARCHAR(30) NOT NULL DEFAULT 'plan_to_watch',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_list_media UNIQUE (list_id, provider_name, provider_media_id, media_type),
    CONSTRAINT chk_watch_status CHECK (watch_status IN ('watched', 'watching', 'plan_to_watch'))
);

CREATE INDEX IF NOT EXISTS idx_custom_list_items_list_id ON custom_list_items(list_id);

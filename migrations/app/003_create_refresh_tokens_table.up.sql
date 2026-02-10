CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token VARCHAR(512) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW() ,
    CREATE INDEX idx_refresh_tokens_user_id ON roles(user_id),
    CREATE INDEX idx_refresh_tokens_token ON roles(token),
    CREATE INDEX idx_refresh_tokens_expires_at ON roles(expires_at),
    CREATE INDEX idx_refresh_tokens_user_revoked ON roles(user_idis_revoked),
    CREATE INDEX idx_refresh_tokens_expires_revoked ON roles(expires_atis_revoked),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) 

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(512) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    CREATE INDEX idx_password_reset_tokens_email ON roles(email),
    CREATE INDEX idx_password_reset_tokens_token ON roles(token),
    CREATE INDEX idx_password_reset_tokens_expires_at ON roles(expires_at),
    CREATE INDEX idx_password_reset_tokens_validation ON roles(emailis_usedexpires_at)
) 

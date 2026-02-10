CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    label VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW() ,
    UNIQUE KEY unique_resource_action (resourceaction),
    CREATE INDEX idx_permissions_resource ON roles(resource),
    CREATE INDEX idx_permissions_action ON roles(action)
) 

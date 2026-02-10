CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    jenis_kelamin VARCHAR(20) NULL,
    foto_profil VARCHAR(500) NULL,
    role_id INTEGER NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW() ,
    CREATE INDEX idx_users_email ON roles(email),
    CREATE INDEX idx_users_is_active ON roles(is_active),
    CREATE INDEX idx_users_role_id ON roles(role_id),
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL
) 

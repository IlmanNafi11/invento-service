CREATE TABLE IF NOT EXISTS moduls (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    nama_file VARCHAR(255) NOT NULL,
    tipe VARCHAR(50) NOT NULL,
    ukuran VARCHAR(50) NOT NULL,
    semester INT NOT NULL DEFAULT 1,
    path_file VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW() ,
    CREATE INDEX idx_moduls_user_id ON roles(user_id),
    CREATE INDEX idx_moduls_tipe ON roles(tipe),
    CREATE INDEX idx_moduls_semester ON roles(semester),
    CREATE INDEX idx_moduls_user_semester ON roles(user_idsemester),
    CREATE INDEX idx_moduls_tipe_semester ON roles(tipesemester),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) 

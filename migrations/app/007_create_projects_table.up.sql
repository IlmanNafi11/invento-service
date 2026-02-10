CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    nama_project VARCHAR(255) NOT NULL,
    kategori VARCHAR(50) NOT NULL,
    semester INT NOT NULL,
    ukuran VARCHAR(50) NOT NULL,
    path_file VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW() ,
    CREATE INDEX idx_projects_user_id ON roles(user_id),
    CREATE INDEX idx_projects_semester ON roles(semester),
    CREATE INDEX idx_projects_kategori ON roles(kategori),
    CREATE INDEX idx_projects_user_semester ON roles(user_idsemester),
    CREATE INDEX idx_projects_kategori_semester ON roles(kategorisemester),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) 

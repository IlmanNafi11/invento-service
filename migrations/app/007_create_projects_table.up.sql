CREATE TABLE IF NOT EXISTS projects (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    nama_project VARCHAR(255) NOT NULL,
    kategori VARCHAR(50) NOT NULL,
    semester INT NOT NULL,
    ukuran VARCHAR(50) NOT NULL,
    path_file VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_projects_user_id (user_id),
    INDEX idx_projects_semester (semester),
    INDEX idx_projects_kategori (kategori),
    INDEX idx_projects_user_semester (user_id, semester),
    INDEX idx_projects_kategori_semester (kategori, semester),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

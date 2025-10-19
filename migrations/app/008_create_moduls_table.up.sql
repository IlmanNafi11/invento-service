CREATE TABLE IF NOT EXISTS moduls (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    nama_file VARCHAR(255) NOT NULL,
    tipe VARCHAR(50) NOT NULL,
    ukuran VARCHAR(50) NOT NULL,
    semester INT NOT NULL DEFAULT 1,
    path_file VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_moduls_user_id (user_id),
    INDEX idx_moduls_tipe (tipe),
    INDEX idx_moduls_semester (semester),
    INDEX idx_moduls_user_semester (user_id, semester),
    INDEX idx_moduls_tipe_semester (tipe, semester),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

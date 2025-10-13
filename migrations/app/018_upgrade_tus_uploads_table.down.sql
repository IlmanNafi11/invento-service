-- Rollback to old structure
DROP TABLE IF EXISTS tus_uploads;

-- Create old structure for rollback  
CREATE TABLE tus_uploads (
    id VARCHAR(36) PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    nama_project VARCHAR(255) NOT NULL,
    kategori VARCHAR(50) NOT NULL,
    semester INT NOT NULL,
    file_size BIGINT NOT NULL,
    current_offset BIGINT DEFAULT 0,
    file_path VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL,
    progress DECIMAL(5,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

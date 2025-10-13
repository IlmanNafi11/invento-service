-- Drop existing table and recreate with new structure
DROP TABLE IF EXISTS tus_uploads;

-- Create table with correct structure
CREATE TABLE tus_uploads (
    id VARCHAR(36) PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    project_id INT UNSIGNED NULL,
    upload_type VARCHAR(20) NOT NULL DEFAULT 'project_create',
    upload_url VARCHAR(500) NULL,
    upload_metadata JSON NULL,
    file_size BIGINT NOT NULL,
    current_offset BIGINT DEFAULT 0,
    file_path VARCHAR(500) NULL,
    status VARCHAR(20) NOT NULL,
    progress DECIMAL(5,2) DEFAULT 0.00,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at),
    INDEX idx_upload_type (upload_type),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE otps ADD COLUMN password_hash VARCHAR(512) DEFAULT '' AFTER user_name;

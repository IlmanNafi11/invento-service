INSERT INTO roles (nama_role) VALUES 
('admin'),
('mahasiswa'),
('dosen')
ON DUPLICATE KEY UPDATE nama_role=VALUES(nama_role);

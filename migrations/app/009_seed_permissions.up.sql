INSERT INTO permissions (resource, action, label) VALUES
('Role', 'create', 'Buat role'),
('Role', 'read', 'Lihat role'),
('Role', 'update', 'Perbarui role'),
('Role', 'delete', 'Hapus role'),
('Permission', 'create', 'Buat permission'),
('Permission', 'read', 'Lihat permission'),
('Permission', 'update', 'Perbarui permission'),
('Permission', 'delete', 'Hapus permission')
ON DUPLICATE KEY UPDATE resource=VALUES(resource);

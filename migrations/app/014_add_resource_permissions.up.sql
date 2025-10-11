INSERT INTO permissions (resource, action, label) VALUES
('User', 'create', 'Buat user'),
('User', 'read', 'Lihat user'),
('User', 'update', 'Perbarui user'),
('User', 'delete', 'Hapus user'),
('Modul', 'create', 'Buat modul'),
('Modul', 'read', 'Lihat modul'),
('Modul', 'update', 'Perbarui modul'),
('Modul', 'delete', 'Hapus modul'),
('Project', 'create', 'Buat project'),
('Project', 'read', 'Lihat project'),
('Project', 'update', 'Perbarui project'),
('Project', 'delete', 'Hapus project')
ON DUPLICATE KEY UPDATE resource=VALUES(resource);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'admin'),
    p.id
FROM permissions p
WHERE p.resource IN ('User', 'Modul', 'Project')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'mahasiswa'),
    p.id
FROM permissions p
WHERE p.resource IN ('Modul', 'Project') AND p.action IN ('create', 'read', 'update', 'delete')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'dosen'),
    p.id
FROM permissions p
WHERE (p.resource IN ('Modul', 'Project') AND p.action IN ('create', 'read', 'update', 'delete'))
   OR (p.resource = 'User' AND p.action = 'read')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

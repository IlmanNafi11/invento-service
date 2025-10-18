INSERT INTO roles (nama_role) VALUES 
('admin'),
('mahasiswa'),
('dosen')
ON DUPLICATE KEY UPDATE nama_role=VALUES(nama_role);

INSERT INTO permissions (resource, action, label) VALUES
('Role', 'create', 'Buat role'),
('Role', 'read', 'Lihat role'),
('Role', 'update', 'Perbarui role'),
('Role', 'delete', 'Hapus role'),
('Permission', 'create', 'Buat permission'),
('Permission', 'read', 'Lihat permission'),
('Permission', 'update', 'Perbarui permission'),
('Permission', 'delete', 'Hapus permission'),
('User', 'create', 'Buat user'),
('User', 'read', 'Lihat user'),
('User', 'update', 'Perbarui user'),
('User', 'delete', 'Hapus user'),
('User', 'download', 'Download data user'),
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
WHERE p.resource IN ('Role', 'Permission', 'User', 'Modul', 'Project')
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
   OR (p.resource = 'User' AND p.action IN ('read', 'download'))
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

INSERT INTO casbin_rule (ptype, v0, v1, v2) 
SELECT DISTINCT 'p', r.nama_role, p.resource, p.action
FROM roles r
INNER JOIN role_permissions rp ON r.id = rp.role_id
INNER JOIN permissions p ON rp.permission_id = p.id
ON DUPLICATE KEY UPDATE v0=VALUES(v0);

INSERT INTO users (email, password, name, is_active, role_id) 
SELECT 
    'admin@admin.polije.ac.id', 
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2', 
    'admin', 
    1,
    id
FROM roles 
WHERE nama_role = 'admin'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@admin.polije.ac.id')
LIMIT 1;

INSERT INTO users (email, password, name, is_active, role_id) 
SELECT 
    'dosen@teacher.polije.ac.id', 
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2', 
    'dosen', 
    1,
    id
FROM roles 
WHERE nama_role = 'dosen'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'dosen@teacher.polije.ac.id')
LIMIT 1;

INSERT INTO users (email, password, name, is_active, role_id) 
SELECT 
    'mahasiswa@student.polije.ac.id', 
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2', 
    'mahasiswa', 
    1,
    id
FROM roles 
WHERE nama_role = 'mahasiswa'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'mahasiswa@student.polije.ac.id')
LIMIT 1;

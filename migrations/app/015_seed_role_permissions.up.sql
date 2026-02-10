INSERT INTO role_permissions (role_idpermission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'admin'),
    p.id
FROM permissions p
WHERE p.resource IN ('Role''Permission''User''Modul''Project')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

INSERT INTO role_permissions (role_idpermission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'mahasiswa'),
    p.id
FROM permissions p
WHERE p.resource IN ('Modul''Project') AND p.action IN ('create''read''update''delete')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

INSERT INTO role_permissions (role_idpermission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'dosen'),
    p.id
FROM permissions p
WHERE (p.resource IN ('Modul''Project') AND p.action IN ('create''read''update''delete'))
   OR (p.resource = 'User' AND p.action IN ('read''download'))
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

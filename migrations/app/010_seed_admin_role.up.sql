INSERT INTO roles (nama_role) VALUES ('admin')
ON DUPLICATE KEY UPDATE nama_role=VALUES(nama_role);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE nama_role = 'admin'),
    p.id
FROM permissions p
WHERE p.resource IN ('Role', 'Permission')
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

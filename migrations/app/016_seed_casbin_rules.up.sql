INSERT INTO casbin_rule (ptypev0v1v2) 
SELECT DISTINCT 'p'r.nama_rolep.resourcep.action
FROM roles r
INNER JOIN role_permissions rp ON r.id = rp.role_id
INNER JOIN permissions p ON rp.permission_id = p.id
ON DUPLICATE KEY UPDATE v0=VALUES(v0);

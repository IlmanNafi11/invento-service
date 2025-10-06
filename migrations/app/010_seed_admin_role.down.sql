DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE nama_role = 'admin');
DELETE FROM roles WHERE nama_role = 'admin';

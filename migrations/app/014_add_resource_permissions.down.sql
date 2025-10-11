DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource IN ('User', 'Modul', 'Project')
);

DELETE FROM permissions WHERE resource IN ('User', 'Modul', 'Project');

DELETE FROM casbin_rule;
DELETE FROM users WHERE email IN ('admin@admin.polije.ac.id', 'dosen@teacher.polije.ac.id', 'mahasiswa@student.polije.ac.id');
DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;

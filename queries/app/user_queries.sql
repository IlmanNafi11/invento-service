-- name: GetUserListWithRole :many
SELECT 
    u.id,
    u.email,
    u.created_at as dibuat_pada,
    COALESCE(r.nama_role, '') as role
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.is_active = true
    AND (? = '' OR u.email LIKE CONCAT('%', ?, '%'))
    AND (? = '' OR r.nama_role = ?)
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountUsersWithSearch :one
SELECT COUNT(*) as total
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.is_active = true
    AND (? = '' OR u.email LIKE CONCAT('%', ?, '%'))
    AND (? = '' OR r.nama_role = ?);

-- name: GetUserFilesFromProjects :many
SELECT 
    p.id,
    p.nama_project as nama_file,
    'Project' as kategori,
    p.path_file as download_url,
    p.updated_at
FROM projects p
WHERE p.user_id = ?
    AND (? = '' OR p.nama_project LIKE CONCAT('%', ?, '%'))
ORDER BY p.updated_at DESC;

-- name: GetUserFilesFromModuls :many
SELECT 
    m.id,
    m.nama_file,
    'Modul' as kategori,
    m.path_file as download_url,
    m.updated_at
FROM moduls m
WHERE m.user_id = ?
    AND (? = '' OR m.nama_file LIKE CONCAT('%', ?, '%'))
ORDER BY m.updated_at DESC;

-- name: CountUserFilesFromProjects :one
SELECT COUNT(*) as total
FROM projects p
WHERE p.user_id = ?
    AND (? = '' OR p.nama_project LIKE CONCAT('%', ?, '%'));

-- name: CountUserFilesFromModuls :one
SELECT COUNT(*) as total
FROM moduls m
WHERE m.user_id = ?
    AND (? = '' OR m.nama_file LIKE CONCAT('%', ?, '%'));

-- name: GetUserProfileStats :one
SELECT 
    COUNT(DISTINCT p.id) as jumlah_project,
    COUNT(DISTINCT m.id) as jumlah_modul
FROM users u
LEFT JOIN projects p ON p.user_id = u.id
LEFT JOIN moduls m ON m.user_id = u.id
WHERE u.id = ? AND u.is_active = true;

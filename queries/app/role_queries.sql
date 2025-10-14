-- Get role list with permission count and search
-- name: GetRoleListWithPermissionCount :many
SELECT 
    r.id,
    r.nama_role,
    COUNT(rp.id) as jumlah_permission,
    r.updated_at as tanggal_diperbarui
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
WHERE ($1 = '' OR r.nama_role LIKE CONCAT('%', $1, '%'))
GROUP BY r.id, r.nama_role, r.updated_at
ORDER BY r.updated_at DESC
LIMIT $2 OFFSET $3;

-- Count total roles with search
-- name: CountRolesWithSearch :one
SELECT COUNT(DISTINCT r.id)
FROM roles r
WHERE ($1 = '' OR r.nama_role LIKE CONCAT('%', $1, '%'));

-- Get role with all permissions details
-- name: GetRoleWithPermissions :many
SELECT 
    r.id as role_id,
    r.nama_role,
    r.created_at as role_created_at,
    r.updated_at as role_updated_at,
    p.id as permission_id,
    p.resource,
    p.action,
    p.label
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
WHERE r.id = $1;

-- Check if role name exists excluding specific ID
-- name: CheckRoleNameExists :one
SELECT COUNT(*) > 0 as exists
FROM roles
WHERE nama_role = $1 AND id != $2;

-- Get all permissions grouped by resource
-- name: GetAvailablePermissionsGrouped :many
SELECT 
    p.resource,
    p.action,
    p.label,
    p.created_at
FROM permissions p
ORDER BY p.resource, p.action;

-- Delete all role permissions for a role
-- name: DeleteRolePermissionsByRoleID :exec
DELETE FROM role_permissions
WHERE role_id = $1;

-- Check if role is used by any user
-- name: CheckRoleUsage :one
SELECT COUNT(*) as usage_count
FROM users
WHERE role_id = $1;

-- Bulk insert role permissions
-- name: BulkInsertRolePermissions :exec
INSERT INTO role_permissions (role_id, permission_id, created_at)
VALUES ($1, $2, NOW());

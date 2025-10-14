-- name: GetProjectsByUserID
-- Mengambil daftar project berdasarkan user ID dengan filtering dan pagination
SELECT 
    id,
    nama_project,
    kategori,
    semester,
    ukuran,
    path_file,
    updated_at as terakhir_diperbarui
FROM projects
WHERE user_id = ?
    AND (? = '' OR nama_project LIKE CONCAT('%', ?, '%'))
    AND (? = 0 OR semester = ?)
    AND (? = '' OR kategori = ?)
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountProjectsByUserID
-- Menghitung total project berdasarkan user ID dengan filtering
SELECT COUNT(*) as total
FROM projects
WHERE user_id = ?
    AND (? = '' OR nama_project LIKE CONCAT('%', ?, '%'))
    AND (? = 0 OR semester = ?)
    AND (? = '' OR kategori = ?);

-- name: GetProjectByID
-- Mengambil detail project berdasarkan ID
SELECT 
    id,
    user_id,
    nama_project,
    kategori,
    semester,
    ukuran,
    path_file,
    created_at,
    updated_at
FROM projects
WHERE id = ?
LIMIT 1;

-- name: GetProjectsByIDs
-- Mengambil multiple projects berdasarkan IDs untuk download
SELECT 
    id,
    user_id,
    nama_project,
    kategori,
    semester,
    ukuran,
    path_file,
    created_at,
    updated_at
FROM projects
WHERE id IN (?)
    AND user_id = ?;

-- name: CountProjectsByUserIDSimple
-- Menghitung total project berdasarkan user ID tanpa filtering
SELECT COUNT(*) as total
FROM projects
WHERE user_id = ?;

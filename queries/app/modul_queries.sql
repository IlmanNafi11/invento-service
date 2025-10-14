-- name: GetModulsByUserID
-- Mengambil daftar modul berdasarkan user ID dengan filtering dan pagination
SELECT 
    id,
    nama_file,
    tipe,
    ukuran,
    semester,
    path_file,
    updated_at as terakhir_diperbarui
FROM moduls
WHERE user_id = ?
    AND (? = '' OR nama_file LIKE CONCAT('%', ?, '%'))
    AND (? = '' OR tipe = ?)
    AND (? = 0 OR semester = ?)
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountModulsByUserID
-- Menghitung total modul berdasarkan user ID dengan filtering
SELECT COUNT(*) as total
FROM moduls
WHERE user_id = ?
    AND (? = '' OR nama_file LIKE CONCAT('%', ?, '%'))
    AND (? = '' OR tipe = ?)
    AND (? = 0 OR semester = ?);

-- name: GetModulByID
-- Mengambil detail modul berdasarkan ID
SELECT 
    id,
    user_id,
    nama_file,
    tipe,
    ukuran,
    semester,
    path_file,
    created_at,
    updated_at
FROM moduls
WHERE id = ?
LIMIT 1;

-- name: GetModulsByIDs
-- Mengambil multiple moduls berdasarkan IDs untuk download
SELECT 
    id,
    user_id,
    nama_file,
    tipe,
    ukuran,
    semester,
    path_file,
    created_at,
    updated_at
FROM moduls
WHERE id IN (?)
    AND user_id = ?;

-- name: CountModulsByUserIDSimple
-- Menghitung total modul berdasarkan user ID tanpa filtering
SELECT COUNT(*) as total
FROM moduls
WHERE user_id = ?;

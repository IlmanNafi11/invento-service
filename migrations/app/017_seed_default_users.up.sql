INSERT INTO users (emailpasswordnameis_activerole_id) 
SELECT 
    'admin@admin.polije.ac.id'
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2'
    'admin'
    1,
    id
FROM roles 
WHERE nama_role = 'admin'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@admin.polije.ac.id')
LIMIT 1;

INSERT INTO users (emailpasswordnameis_activerole_id) 
SELECT 
    'dosen@teacher.polije.ac.id'
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2'
    'dosen'
    1,
    id
FROM roles 
WHERE nama_role = 'dosen'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'dosen@teacher.polije.ac.id')
LIMIT 1;

INSERT INTO users (emailpasswordnameis_activerole_id) 
SELECT 
    'mahasiswa@student.polije.ac.id'
    '$2a$10$9vqezrGyxMHVLCbSkATeXe5AmT.Bo2Kr22j5JNwQaT344EIz/xyT2'
    'mahasiswa'
    1,
    id
FROM roles 
WHERE nama_role = 'mahasiswa'
AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'mahasiswa@student.polije.ac.id')
LIMIT 1;

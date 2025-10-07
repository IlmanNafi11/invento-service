INSERT INTO roles (nama_role) VALUES ('mahasiswa')
ON DUPLICATE KEY UPDATE nama_role=VALUES(nama_role);

INSERT INTO roles (nama_role) VALUES ('dosen')
ON DUPLICATE KEY UPDATE nama_role=VALUES(nama_role);

package config

import (
	"fiber-boiler-plate/internal/domain"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB, cfg *Config) {
	log.Println("Menjalankan database seeder...")

	seedPermissions(db)
	seedAdminRole(db)
	seedMahasiswaDosenRoles(db)

	if cfg.Database.SeedUsers {
		seedUsers(db)
	} else {
		log.Println("Seeder users dinonaktifkan melalui konfigurasi")
	}

	assignRoleToExistingUsers(db)

	log.Println("Database seeder selesai")
}

func seedUsers(db *gorm.DB) {
	var count int64
	db.Model(&domain.User{}).Where("email = ?", "admin@admin.polije.ac.id").Count(&count)

	if count > 0 {
		log.Println("Admin user seed sudah ada, melewati seeding user")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("polije"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Gagal hash password:", err)
	}

	var adminRole domain.Role
	if err := db.Where("nama_role = ?", "admin").First(&adminRole).Error; err != nil {
		log.Fatal("Admin role tidak ditemukan:", err)
	}

	user := domain.User{
		Email:    "admin@admin.polije.ac.id",
		Password: string(hashedPassword),
		Name:     "Administrator",
		RoleID:   &adminRole.ID,
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Gagal membuat admin user seed:", err)
	}

	log.Println("Admin user seed berhasil dibuat dengan email: admin@admin.polije.ac.id")
}

func seedPermissions(db *gorm.DB) {
	permissions := []domain.Permission{
		{Resource: "Role", Action: "create", Label: "Buat role"},
		{Resource: "Role", Action: "read", Label: "Lihat role"},
		{Resource: "Role", Action: "update", Label: "Perbarui role"},
		{Resource: "Role", Action: "delete", Label: "Hapus role"},
		{Resource: "Permission", Action: "create", Label: "Buat permission"},
		{Resource: "Permission", Action: "read", Label: "Lihat permission"},
		{Resource: "Permission", Action: "update", Label: "Perbarui permission"},
		{Resource: "Permission", Action: "delete", Label: "Hapus permission"},
		{Resource: "Project", Action: "create", Label: "Buat project"},
		{Resource: "Project", Action: "update", Label: "Perbarui project"},
		{Resource: "Project", Action: "read", Label: "Lihat project"},
		{Resource: "Project", Action: "delete", Label: "Hapus project"},
		{Resource: "Modul", Action: "create", Label: "Buat modul"},
		{Resource: "Modul", Action: "update", Label: "Perbarui modul"},
		{Resource: "Modul", Action: "read", Label: "Lihat modul"},
		{Resource: "Modul", Action: "delete", Label: "Hapus modul"},
		{Resource: "User", Action: "update", Label: "Perbarui user"},
		{Resource: "User", Action: "read", Label: "Lihat user"},
		{Resource: "User", Action: "delete", Label: "Hapus user"},
	}

	var createdCount int
	for _, perm := range permissions {
		var existing domain.Permission
		result := db.Where("resource = ? AND action = ?", perm.Resource, perm.Action).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				log.Printf("Gagal membuat permission %s %s: %v", perm.Resource, perm.Action, err)
			} else {
				createdCount++
			}
		}
	}

	if createdCount > 0 {
		log.Printf("Permissions seed berhasil dibuat: %d permissions baru", createdCount)
	} else {
		log.Println("Tidak ada permission baru yang perlu dibuat")
	}
}

func seedAdminRole(db *gorm.DB) {
	var count int64
	db.Model(&domain.Role{}).Where("nama_role = ?", "admin").Count(&count)

	if count > 0 {
		log.Println("Admin role sudah ada, melewati seeding admin role")
		return
	}

	adminRole := domain.Role{
		NamaRole: "admin",
	}

	if err := db.Create(&adminRole).Error; err != nil {
		log.Fatal("Gagal membuat admin role:", err)
	}

	var permissions []domain.Permission
	if err := db.Find(&permissions).Error; err != nil {
		log.Fatal("Gagal mengambil permissions:", err)
	}

	var rolePermissions []domain.RolePermission
	for _, perm := range permissions {
		rolePermissions = append(rolePermissions, domain.RolePermission{
			RoleID:       adminRole.ID,
			PermissionID: perm.ID,
		})
	}

	if err := db.Create(&rolePermissions).Error; err != nil {
		log.Fatal("Gagal membuat role_permissions:", err)
	}

	log.Printf("Admin role seed berhasil dibuat dengan %d permissions", len(rolePermissions))
}

func seedMahasiswaDosenRoles(db *gorm.DB) {
	var mahasiswaCount int64
	db.Model(&domain.Role{}).Where("nama_role = ?", "mahasiswa").Count(&mahasiswaCount)

	var dosenCount int64
	db.Model(&domain.Role{}).Where("nama_role = ?", "dosen").Count(&dosenCount)

	var mahasiswaRole domain.Role
	if mahasiswaCount == 0 {
		mahasiswaRole = domain.Role{
			NamaRole: "mahasiswa",
		}
		if err := db.Create(&mahasiswaRole).Error; err != nil {
			log.Fatal("Gagal membuat role mahasiswa:", err)
		}
		log.Println("Role mahasiswa seed berhasil dibuat")
	} else {
		if err := db.Where("nama_role = ?", "mahasiswa").First(&mahasiswaRole).Error; err != nil {
			log.Fatal("Gagal mengambil role mahasiswa:", err)
		}
	}

	var dosenRole domain.Role
	if dosenCount == 0 {
		dosenRole = domain.Role{
			NamaRole: "dosen",
		}
		if err := db.Create(&dosenRole).Error; err != nil {
			log.Fatal("Gagal membuat role dosen:", err)
		}
		log.Println("Role dosen seed berhasil dibuat")
	} else {
		if err := db.Where("nama_role = ?", "dosen").First(&dosenRole).Error; err != nil {
			log.Fatal("Gagal mengambil role dosen:", err)
		}
	}

	assignPermissionsToRole(db, mahasiswaRole.ID, []string{"Modul", "Project"})
	assignPermissionsToRole(db, dosenRole.ID, []string{"Modul", "Project", "User"})
}

func assignPermissionsToRole(db *gorm.DB, roleID uint, resources []string) {
	for _, resource := range resources {
		var permissions []domain.Permission
		if err := db.Where("resource = ?", resource).Find(&permissions).Error; err != nil {
			log.Printf("Gagal mengambil permissions untuk resource %s: %v", resource, err)
			continue
		}

		for _, perm := range permissions {
			rolePermission := &domain.RolePermission{
				RoleID:       roleID,
				PermissionID: perm.ID,
			}

			var count int64
			db.Model(&domain.RolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, perm.ID).Count(&count)
			if count == 0 {
				if err := db.Create(rolePermission).Error; err != nil {
					log.Printf("Gagal assign permission %s %s ke role: %v", resource, perm.Action, err)
				}
			}
		}
	}
}

func assignRoleToExistingUsers(db *gorm.DB) {
	var adminRole domain.Role
	if err := db.Where("nama_role = ?", "admin").First(&adminRole).Error; err != nil {
		log.Printf("Admin role tidak ditemukan, skip assign role ke existing users: %v", err)
		return
	}

	result := db.Model(&domain.User{}).Where("email = ? AND role_id IS NULL", "user@example.com").Update("role_id", adminRole.ID)
	if result.Error != nil {
		log.Printf("Gagal assign role ke user@example.com: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Println("Berhasil assign role admin ke user@example.com")
	}
}

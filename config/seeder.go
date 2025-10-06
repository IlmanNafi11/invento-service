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

	if cfg.Database.SeedUsers {
		seedUsers(db)
	} else {
		log.Println("Seeder users dinonaktifkan melalui konfigurasi")
	}

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
	var count int64
	db.Model(&domain.Permission{}).Count(&count)

	if count > 0 {
		log.Println("Permissions sudah ada, melewati seeding permissions")
		return
	}

	permissions := []domain.Permission{
		{Resource: "Role", Action: "create", Label: "Buat role"},
		{Resource: "Role", Action: "read", Label: "Lihat role"},
		{Resource: "Role", Action: "update", Label: "Perbarui role"},
		{Resource: "Role", Action: "delete", Label: "Hapus role"},
		{Resource: "Permission", Action: "create", Label: "Buat permission"},
		{Resource: "Permission", Action: "read", Label: "Lihat permission"},
		{Resource: "Permission", Action: "update", Label: "Perbarui permission"},
		{Resource: "Permission", Action: "delete", Label: "Hapus permission"},
	}

	if err := db.Create(&permissions).Error; err != nil {
		log.Fatal("Gagal membuat permissions seed:", err)
	}

	log.Printf("Permissions seed berhasil dibuat: %d permissions", len(permissions))
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
	if err := db.Where("resource IN ?", []string{"Role", "Permission"}).Find(&permissions).Error; err != nil {
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

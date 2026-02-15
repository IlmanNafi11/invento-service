package main

import (
	"invento-service/config"
)

// @title                       Invento Service API
// @version                     1.0
// @description                 API untuk Invento Service - Sistem manajemen proyek dan modul yang lengkap.
// @description.name            API ini menyediakan endpoint untuk autentikasi, manajemen pengguna, proyek, modul, dan statistik.
// @description.name            Mendukung autentikasi JWT, upload file dengan protokol TUS, dan kontrol akses berbasis peran (RBAC).
//
// @contact.name                Invento Support
// @contact.url                 https://github.com/ilmannafi
// @contact.email               support@invento.com
//
// @license.name                MIT
// @license.url                 https://opensource.org/licenses/MIT
//
// @host                        localhost:3000
// @BasePath                    /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Token JWT dengan format "Bearer {token}". Refresh token dikirim melalui httpOnly cookie.
func init() {
	// This file is used by swag to generate API documentation
	_ = config.Config{}
}

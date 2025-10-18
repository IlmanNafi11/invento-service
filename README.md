# Fiber Boilerplate

Backend boilerplate menggunakan Fiber Framework dengan implementasi Clean Architecture, authentication system, dan Redis caching.

## Fitur

- ✅ Clean Architecture dengan Domain Driven Design
- ✅ Authentication System (Register, Login, Refresh Token, Reset Password, Logout)
- ✅ PostgreSQL Integration dengan GORM
- ✅ Redis Integration dengan Health Check & Monitoring
- ✅ JWT Token Authentication dengan Access & Refresh Token
- ✅ Password Hashing dengan BCrypt
- ✅ Request Validation dengan Comprehensive Error Handling
- ✅ Standardized API Response dengan Pagination Support
- ✅ Database Migration & Seeder Otomatis
- ✅ Health Check & Monitoring System (4 endpoints)
- ✅ Middleware Support (Auth, CORS, Logger, Recovery)
- ✅ Environment Configuration dengan Validasi
- ✅ OpenAPI Documentation (Swagger)
- ✅ Comprehensive Testing Suite (90+ tests)
- ✅ Redis Repository Pattern dengan JSON Support
- ✅ Graceful Shutdown & Error Handling

## Stack Teknologi

- **Bahasa**: Go 1.24.0
- **Framework**: Fiber v2.52.9
- **Database**: PostgreSQL dengan GORM v1.25.12
- **Cache**: Redis v9.14.0 dengan go-redis
- **Authentication**: JWT v4.5.2
- **Validation**: Go Playground Validator v10.27.0
- **Config**: Viper v1.21.0
- **Password Hash**: BCrypt
- **UUID**: Google UUID v1.6.0
- **Testing**: Go Testing dengan Testify

## Struktur Project

```
fiber-boiler-plate/
├── api/                       # OpenAPI specifications
│   ├── auth.yaml             # Authentication endpoints
│   └── health.yaml           # Health check endpoints
├── cmd/app/                  # Entry point aplikasi
│   └── main.go              # Main application dengan dependency injection
├── config/                   # Konfigurasi aplikasi
│   ├── config.go            # Konfigurasi utama
│   ├── database.go          # Konfigurasi database
│   ├── redis.go             # Konfigurasi Redis dengan graceful degradation
│   └── test/                # Unit tests untuk konfigurasi
├── docs/                     # Dokumentasi
│   └── api-response-standard.md
├── gen/app/db/              # Generated database files
├── internal/                 # Kode internal aplikasi
│   ├── app/                 # Setup server & middleware
│   │   └── server.go        # HTTP server configuration
│   ├── controller/http/     # HTTP handlers/controllers
│   │   ├── auth_controller.go
│   │   ├── health_controller.go
│   │   └── test/            # Controller tests
│   ├── domain/              # Domain entities & models
│   │   ├── auth.go          # Authentication entities
│   │   ├── health.go        # Health check entities dengan Redis
│   │   ├── response.go      # Response models
│   │   ├── user.go          # User entities
│   │   └── test/            # Domain tests
│   ├── helper/              # Utilities & helpers
│   │   ├── jwt.go           # JWT utilities
│   │   ├── response.go      # Response helpers
│   │   ├── validation.go    # Validation helpers
│   │   └── test/            # Helper tests
│   └── usecase/             # Business logic
│       ├── auth_usecase.go  # Authentication use cases
│       ├── health_usecase.go # Health check use cases
│       ├── repo/            # Repository layer
│       │   ├── password_reset_token_repository.go
│       │   ├── redis_repository.go  # Redis repository implementation
│       │   ├── refresh_token_repository.go
│       │   ├── user_repository.go
│       │   └── test/        # Repository tests
│       └── test/            # Use case tests
├── migrations/app/          # Database migrations
│   ├── 001_create_users_table.sql
│   ├── 002_create_refresh_tokens_table.sql
│   └── 003_create_password_reset_tokens_table.sql
├── queries/app/             # SQL queries
├── .env                     # Environment variables (gitignored)
├── .env.example            # Environment template
├── go.mod                  # Go modules
├── go.sum                  # Dependencies checksum
└── main.go                 # Application entry point
```

## Setup Development

### 1. Prerequisites

- Go 1.24+
- PostgreSQL 13+
- Redis 6.0+ (Optional, aplikasi tetap berjalan tanpa Redis)
- Git

### 2. Clone Repository

```bash
git clone <repository-url>
cd fiber-boiler-plate
```

### 3. Environment Setup

Copy file `.env.example` ke `.env`:

```bash
cp .env.example .env
```

Edit file `.env` sesuai konfigurasi environment Anda:

```env
# Application Configuration
APP_ENV=development
APP_NAME=Fiber Boilerplate
APP_PORT=3000

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=fiber_boilerplate
DB_SSL_MODE=disable
DB_AUTO_MIGRATE=true
DB_SEED_DATA=true

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRE_HOURS=24
REFRESH_TOKEN_EXPIRE_HOURS=168

# Mail Configuration (untuk reset password)
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-app-password
MAIL_FROM=noreply@yourapp.com

# Redis Configuration (Optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
REDIS_POOL_TIMEOUT=30
REDIS_IDLE_TIMEOUT=300
REDIS_IDLE_CHECK_FREQUENCY=60
```

### 4. Database Setup

Buat database PostgreSQL:

```sql
CREATE DATABASE fiber_boilerplate;
```

### 5. Redis Setup (Optional)

Install dan jalankan Redis server:

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

**macOS:**
```bash
brew install redis
brew services start redis
```

**Docker:**
```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

### 6. Install Dependencies

```bash
go mod tidy
```

### 7. Run Application

**Development mode:**
```bash
go run main.go
```

**Build dan jalankan:**
```bash
go build -o bin/app ./main.go
./bin/app
```

**Menggunakan go run dengan cmd/app:**
```bash
go run cmd/app/main.go
```

Server akan berjalan di http://localhost:3000

## API Endpoints

### Authentication

| Method | Endpoint | Deskripsi | Auth Required |
|--------|----------|-----------|---------------|
| POST | `/api/v1/auth/register` | Registrasi user baru | ❌ |
| POST | `/api/v1/auth/login` | Login user | ❌ |
| POST | `/api/v1/auth/refresh` | Refresh access token | ❌ |
| POST | `/api/v1/auth/reset-password` | Request reset password | ❌ |
| POST | `/api/v1/auth/reset-password/confirm` | Konfirmasi reset password | ❌ |
| POST | `/api/v1/auth/logout` | Logout user | ✅ |

### Health Check & Monitoring

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/health` | Basic health check |
| GET | `/api/v1/monitoring/health` | Comprehensive health check (includes Redis) |
| GET | `/api/v1/monitoring/metrics` | System metrics (includes Redis metrics) |
| GET | `/api/v1/monitoring/status` | Application status (includes Redis service status) |

## Contoh Request & Response

### Register User
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Health Check dengan Redis
```bash
curl http://localhost:3000/api/v1/monitoring/health
```

Response:
```json
{
  "success": true,
  "message": "Pemeriksaan kesehatan sistem berhasil",
  "code": 200,
  "data": {
    "status": "healthy",
    "app": {
      "name": "fiber-boiler-plate",
      "version": "1.0.0",
      "environment": "development",
      "uptime": "5h 30m 15s"
    },
    "database": {
      "status": "connected",
      "ping_time": "2ms",
      "open_connections": 5,
      "max_connections": 100
    },
    "redis": {
      "status": "connected",
      "ping_time": "1ms",
      "connected_clients": 2,
      "used_memory": "1.2MB",
      "keyspace_hits": 1500,
      "keyspace_misses": 50
    },
    "system": {
      "memory_usage": "45.2MB",
      "cpu_cores": 4,
      "goroutines": 12
    }
  }
}
```

## Redis Integration

### Redis Repository Pattern

Redis terintegrasi menggunakan repository pattern dengan fitur:

- **Set/Get** dengan TTL support
- **JSON serialization** untuk complex objects
- **Increment/Decrement** untuk counters
- **Key expiration** management
- **Pipeline** support untuk batch operations
- **Health monitoring** dengan comprehensive metrics

### Redis Configuration

Redis dikonfigurasi dengan graceful degradation:

- Aplikasi tetap berjalan jika Redis tidak tersedia
- Health check melaporkan status Redis secara terpisah
- Connection pooling dengan retry mechanism
- Timeout configuration untuk reliability

### Redis Health Metrics

Health check Redis menyediakan metrics:
- Connection status dan ping time
- Memory usage (used/max)
- Client connections
- Keyspace hits/misses ratio
- Total commands processed
- Redis server version

## API Documentation

Dokumentasi lengkap tersedia dalam format OpenAPI 3.0:

- **Authentication**: `api/auth.yaml`
- **Health Check**: `api/health.yaml`

## Default User

Aplikasi akan otomatis membuat user default saat pertama kali dijalankan:

- **Email**: user@example.com
- **Password**: user1234
- **Name**: user example

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Pesan sukses",
  "code": 200,
  "data": {
    // data response
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Pesan error",
  "code": 400,
  "errors": {
    // detail error (opsional)
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Validation Error Response
```json
{
  "success": false,
  "message": "Validasi gagal",
  "code": 422,
  "errors": {
    "field_name": ["Error message 1", "Error message 2"]
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Development Guidelines

- Mengikuti Clean Architecture principles
- Mengikuti DRY principle
- Konsistensi naming convention dan struktur
- Redis operations menggunakan repository pattern
- Error handling dengan graceful degradation

## Testing

Jalankan semua tests:
```bash
go test ./...
```

Jalankan tests dengan coverage:
```bash
go test -cover ./...
```

Jalankan tests secara verbose:
```bash
go test -v ./...
```

Test specific package:
```bash
go test ./internal/usecase/repo/test -v
go test ./config/test -v
go test ./internal/domain/test -v
```

## Build untuk Production

```bash
go build -ldflags="-s -w" -o bin/app ./cmd/app/main.go
```

Atau menggunakan main.go di root:
```bash
go build -ldflags="-s -w" -o bin/app ./main.go
```

## Performance & Monitoring

- Health check endpoints untuk monitoring
- Redis metrics untuk cache performance
- Database connection pooling
- JWT token dengan appropriate expiration
- Request validation untuk security
- Structured logging untuk debugging

## Security Features

- Password hashing dengan BCrypt
- JWT token authentication
- Refresh token rotation
- Request validation
- SQL injection prevention dengan GORM
- CORS configuration
- Rate limiting ready (middleware support)

## CI/CD Pipeline

Project ini dilengkapi dengan pipeline CI/CD komprehensif menggunakan GitHub Actions:

### Continuous Integration (CI)
- **Lint**: Code quality checks dengan golangci-lint
- **Test**: Unit tests dengan code coverage minimum 40%
- **Build**: Automatic binary compilation
- **Security**: Vulnerability scanning dengan Gosec

### Continuous Deployment (CD)
- **Docker**: Multi-stage build untuk optimisasi ukuran image
- **Registry**: Push otomatis ke GitHub Container Registry
- **Deployment**: Auto-deployment ke Dev/Staging dari push, Production dari tags

### Local Development

```bash
# Setup development environment
make install-tools
make deps

# Build dan test
make build
make test
make coverage

# Linting dan formatting
make lint
make fmt

# Docker development
docker-compose up -d
docker-compose logs -f app
```

Untuk dokumentasi lengkap, lihat [CI/CD Pipeline Documentation](./docs/CI_CD_PIPELINE.md)

## Deployment

Aplikasi dapat anda deploy ke:
- Docker containers
- Cloud platforms (AWS, GCP, Azure)
- Kubernetes

Environment variables harus dikonfigurasi sesuai dengan production requirements.

## Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

Pastikan semua tests passed dan mengikuti coding standards.

## License

MIT License# CI/CD Test Push

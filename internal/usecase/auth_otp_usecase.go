package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"time"

	"gorm.io/gorm"
)

type AuthOTPUsecase interface {
	RegisterWithOTP(req domain.RegisterRequest) (*domain.OTPResponse, error)
	VerifyRegisterOTP(req domain.VerifyOTPRequest) (string, *domain.AuthResponse, error)
	ResendRegisterOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error)
	InitiateResetPassword(req domain.ResetPasswordRequest) (*domain.OTPResponse, error)
	VerifyResetPasswordOTP(req domain.VerifyOTPRequest) (*domain.OTPResponse, error)
	ConfirmResetPasswordWithOTP(email string, newPassword string) (string, *domain.AuthResponse, error)
	ResendResetPasswordOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error)
}

type authOTPUsecase struct {
	userRepo         repo.UserRepository
	refreshTokenRepo repo.RefreshTokenRepository
	otpRepo          repo.OTPRepository
	roleRepo         repo.RoleRepository
	authHelper       *helper.AuthHelper
	jwtManager       *helper.JWTManager
	mailtrapClient   *helper.MailtrapClient
	otpValidator     *helper.OTPValidator
	rateLimiter      *helper.ResendRateLimiter
	config           *config.Config
	logger           *helper.Logger
}

func NewAuthOTPUsecase(
	userRepo repo.UserRepository,
	refreshTokenRepo repo.RefreshTokenRepository,
	otpRepo repo.OTPRepository,
	roleRepo repo.RoleRepository,
	config *config.Config,
) AuthOTPUsecase {
	jwtManager, err := helper.NewJWTManager(config)
	if err != nil {
		panic("Gagal inisialisasi JWT Manager: " + err.Error())
	}

	authHelper := helper.NewAuthHelper(refreshTokenRepo, jwtManager, config)
	mailtrapClient := helper.NewMailtrapClient(&config.Mailtrap)
	otpValidator := helper.NewOTPValidator(config.OTP.MaxAttempts, config.OTP.ExpiryMinutes)
	rateLimiter := helper.NewResendRateLimiter(config.OTP.ResendMaxTimes, config.OTP.ResendCooldownSeconds)
	logger := helper.NewLogger()

	return &authOTPUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		otpRepo:          otpRepo,
		roleRepo:         roleRepo,
		authHelper:       authHelper,
		jwtManager:       jwtManager,
		mailtrapClient:   mailtrapClient,
		otpValidator:     otpValidator,
		rateLimiter:      rateLimiter,
		config:           config,
		logger:           logger,
	}
}

func (uc *authOTPUsecase) RegisterWithOTP(req domain.RegisterRequest) (*domain.OTPResponse, error) {
	emailInfo, err := helper.ValidatePolijeEmail(req.Email)
	if err != nil {
		return nil, err
	}

	existingUser, _ := uc.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	otp, err := helper.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return nil, errors.New("gagal generate kode otp")
	}

	otpHash := helper.HashOTP(otp)
	expiresAt := time.Now().Add(time.Duration(uc.config.OTP.ExpiryMinutes) * time.Minute)

	otpRecord, err := uc.otpRepo.Create(req.Email, req.Name, otpHash, domain.OTPTypeRegister, expiresAt, uc.config.OTP.MaxAttempts)
	if err != nil {
		return nil, errors.New("gagal menyimpan kode otp")
	}

	if err := uc.mailtrapClient.SendOTPEmail(req.Email, otp, uc.config.Mailtrap.TemplateRegister); err != nil {
		uc.otpRepo.MarkAsUsed(otpRecord.ID)
		return nil, errors.New("gagal mengirim kode otp ke email")
	}

	_ = emailInfo

	return &domain.OTPResponse{
		Message:   "Kode OTP telah dikirim ke email Anda",
		ExpiresIn: uc.config.OTP.ExpiryMinutes * 60,
	}, nil
}

func (uc *authOTPUsecase) VerifyRegisterOTP(req domain.VerifyOTPRequest) (string, *domain.AuthResponse, error) {
	otpType := domain.OTPType(req.Type)
	otpRecord, err := uc.otpRepo.GetByEmail(req.Email, otpType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("kode otp tidak valid atau sudah expired")
		}
		return "", nil, errors.New("gagal mengambil data kode otp")
	}

	if err := uc.otpValidator.ValidateOTPRecord(otpRecord.Attempts, otpRecord.MaxAttempts); err != nil {
		return "", nil, err
	}

	if !helper.VerifyOTPHash(req.Code, otpRecord.CodeHash) {
		uc.otpRepo.IncrementAttempts(otpRecord.ID)
		return "", nil, errors.New("kode otp salah, coba lagi")
	}

	emailInfo, err := helper.ValidatePolijeEmail(req.Email)
	if err != nil {
		return "", nil, err
	}

	role, err := uc.roleRepo.GetByName(emailInfo.RoleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("role tidak tersedia, silakan hubungi administrator")
		}
		return "", nil, errors.New("gagal mengambil data role")
	}

	hashedPassword, err := helper.HashPassword(req.Email + "temp")
	if err != nil {
		return "", nil, err
	}

	user := &domain.User{
		Name:     otpRecord.UserName,
		Email:    req.Email,
		Password: hashedPassword,
		RoleID:   &role.ID,
		IsActive: true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return "", nil, errors.New("gagal membuat user")
	}

	user.Role = role

	if err := uc.otpRepo.MarkAsUsed(otpRecord.ID); err != nil {
		return "", nil, errors.New("gagal menandai kode otp sebagai sudah digunakan")
	}

	if err := uc.otpRepo.DeleteByEmail(req.Email, otpType); err != nil {
		return "", nil, errors.New("gagal menghapus kode otp")
	}

	refreshToken, authResp, err := uc.authHelper.GenerateAuthResponse(user)
	if err != nil {
		return "", nil, err
	}

	return refreshToken, authResp, nil
}

func (uc *authOTPUsecase) ResendRegisterOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error) {
	otpRecord, err := uc.otpRepo.GetByEmail(req.Email, domain.OTPTypeRegister)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("silakan daftar ulang, sesi OTP telah berakhir")
		}
		return nil, errors.New("gagal mengambil data kode otp")
	}

	can, _ := uc.rateLimiter.CanResend(otpRecord.ResendCount, otpRecord.LastResendAt)
	if !can {
		return nil, errors.New("silakan tunggu beberapa saat sebelum mengirim ulang kode otp")
	}

	otp, err := helper.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return nil, errors.New("gagal generate kode otp")
	}

	otpHash := helper.HashOTP(otp)
	expiresAt := time.Now().Add(time.Duration(uc.config.OTP.ExpiryMinutes) * time.Minute)

	newOTPRecord, err := uc.otpRepo.Create(req.Email, otpRecord.UserName, otpHash, domain.OTPTypeRegister, expiresAt, uc.config.OTP.MaxAttempts)
	if err != nil {
		return nil, errors.New("gagal menyimpan kode otp")
	}

	if err := uc.mailtrapClient.SendOTPEmail(req.Email, otp, uc.config.Mailtrap.TemplateRegister); err != nil {
		uc.otpRepo.MarkAsUsed(newOTPRecord.ID)
		return nil, errors.New("gagal mengirim kode otp ke email")
	}

	uc.otpRepo.DeleteByEmail(req.Email, domain.OTPTypeRegister)
	uc.otpRepo.UpdateResendInfo(newOTPRecord.ID, otpRecord.ResendCount+1, time.Now())

	return &domain.OTPResponse{
		Message:   "Kode OTP baru telah dikirim ke email Anda",
		ExpiresIn: uc.config.OTP.ExpiryMinutes * 60,
	}, nil
}

func (uc *authOTPUsecase) InitiateResetPassword(req domain.ResetPasswordRequest) (*domain.OTPResponse, error) {
	// Step 1: Validate user exists
	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	// Step 2: Generate OTP
	otp, err := helper.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return nil, errors.New("gagal generate kode otp")
	}

	// Step 3: Hash OTP
	otpHash := helper.HashOTP(otp)

	// Step 4: Calculate expiry
	expiresAt := time.Now().Add(time.Duration(uc.config.OTP.ExpiryMinutes) * time.Minute)

	// Step 5: Save OTP to database
	otpRecord, err := uc.otpRepo.Create(req.Email, "", otpHash, domain.OTPTypeResetPassword, expiresAt, uc.config.OTP.MaxAttempts)
	if err != nil {
		return nil, errors.New("gagal menyimpan kode otp")
	}

	// Step 6: Send email via Mailtrap
	if err := uc.mailtrapClient.SendOTPEmail(req.Email, otp, uc.config.Mailtrap.TemplateResetPass); err != nil {
		uc.otpRepo.MarkAsUsed(otpRecord.ID)
		return nil, errors.New("gagal mengirim kode otp ke email")
	}

	_ = user

	return &domain.OTPResponse{
		Message:   "Kode OTP untuk reset password telah dikirim ke email Anda",
		ExpiresIn: uc.config.OTP.ExpiryMinutes * 60,
	}, nil
}

func (uc *authOTPUsecase) VerifyResetPasswordOTP(req domain.VerifyOTPRequest) (*domain.OTPResponse, error) {
	otpType := domain.OTPType(req.Type)

	otpRecord, err := uc.otpRepo.GetByEmail(req.Email, otpType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kode otp tidak valid atau sudah expired")
		}
		return nil, errors.New("gagal mengambil data kode otp")
	}

	if err := uc.otpValidator.ValidateOTPRecord(otpRecord.Attempts, otpRecord.MaxAttempts); err != nil {
		return nil, err
	}

	if !helper.VerifyOTPHash(req.Code, otpRecord.CodeHash) {
		uc.otpRepo.IncrementAttempts(otpRecord.ID)
		return nil, errors.New("kode otp salah, coba lagi")
	}

	if err := uc.otpRepo.MarkAsUsed(otpRecord.ID); err != nil {
		return nil, errors.New("gagal menandai kode otp sebagai sudah digunakan")
	}

	return &domain.OTPResponse{
		Message:   "Kode OTP valid, silakan masukkan password baru Anda",
		ExpiresIn: uc.config.OTP.ExpiryMinutes * 60,
	}, nil
}

func (uc *authOTPUsecase) ConfirmResetPasswordWithOTP(email string, newPassword string) (string, *domain.AuthResponse, error) {
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("user tidak ditemukan")
		}
		return "", nil, errors.New("gagal mengambil data user")
	}

	hashedPassword, err := helper.HashPassword(newPassword)
	if err != nil {
		return "", nil, err
	}

	if err := uc.userRepo.UpdatePassword(email, hashedPassword); err != nil {
		return "", nil, errors.New("gagal update password")
	}

	if err := uc.otpRepo.DeleteByEmail(email, domain.OTPTypeResetPassword); err != nil {
		return "", nil, errors.New("gagal menghapus kode otp")
	}

	user.Role = &domain.Role{
		ID:       *user.RoleID,
		NamaRole: "User",
	}

	refreshToken, authResp, err := uc.authHelper.GenerateAuthResponse(user)
	if err != nil {
		return "", nil, err
	}

	return refreshToken, authResp, nil
}

func (uc *authOTPUsecase) ResendResetPasswordOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error) {
	otpRecord, err := uc.otpRepo.GetByEmail(req.Email, domain.OTPTypeResetPassword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("silakan mulai ulang proses reset password")
		}
		return nil, errors.New("gagal mengambil data kode otp")
	}

	can, _ := uc.rateLimiter.CanResend(otpRecord.ResendCount, otpRecord.LastResendAt)
	if !can {
		return nil, errors.New("silakan tunggu beberapa saat sebelum mengirim ulang kode otp")
	}

	otp, err := helper.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return nil, errors.New("gagal generate kode otp")
	}

	otpHash := helper.HashOTP(otp)
	expiresAt := time.Now().Add(time.Duration(uc.config.OTP.ExpiryMinutes) * time.Minute)

	newOTPRecord, err := uc.otpRepo.Create(req.Email, "", otpHash, domain.OTPTypeResetPassword, expiresAt, uc.config.OTP.MaxAttempts)
	if err != nil {
		return nil, errors.New("gagal menyimpan kode otp")
	}

	if err := uc.mailtrapClient.SendOTPEmail(req.Email, otp, uc.config.Mailtrap.TemplateResetPass); err != nil {
		uc.otpRepo.MarkAsUsed(newOTPRecord.ID)
		return nil, errors.New("gagal mengirim kode otp ke email")
	}

	uc.otpRepo.DeleteByEmail(req.Email, domain.OTPTypeResetPassword)
	uc.otpRepo.UpdateResendInfo(newOTPRecord.ID, otpRecord.ResendCount+1, time.Now())

	return &domain.OTPResponse{
		Message:   "Kode OTP baru telah dikirim ke email Anda",
		ExpiresIn: uc.config.OTP.ExpiryMinutes * 60,
	}, nil
}

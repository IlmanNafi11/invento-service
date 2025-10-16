package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fiber-boiler-plate/config"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	RoleID    *uint  `json:"role_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	KeyID     string `json:"kid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	privateKeyRotation *rsa.PrivateKey
	publicKeyRotation  *rsa.PublicKey
	expireHours        int
	useRotationKey     bool
}

func NewJWTManager(cfg *config.Config) (*JWTManager, error) {
	privateKey, err := loadPrivateKey(cfg.JWT.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := loadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyRotation, err := loadPrivateKey(cfg.JWT.PrivateKeyRotationPath)
	if err != nil {
		return nil, err
	}

	publicKeyRotation, err := loadPublicKey(cfg.JWT.PublicKeyRotationPath)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		privateKey:         privateKey,
		publicKey:          publicKey,
		privateKeyRotation: privateKeyRotation,
		publicKeyRotation:  publicKeyRotation,
		expireHours:        cfg.JWT.ExpireHours,
		useRotationKey:     false,
	}, nil
}

func (jm *JWTManager) RotateKeys() {
	jm.useRotationKey = !jm.useRotationKey
}

func (jm *JWTManager) GetCurrentPrivateKey() *rsa.PrivateKey {
	if jm.useRotationKey {
		return jm.privateKeyRotation
	}
	return jm.privateKey
}

func (jm *JWTManager) GetCurrentPublicKey() *rsa.PublicKey {
	if jm.useRotationKey {
		return jm.publicKeyRotation
	}
	return jm.publicKey
}

func (jm *JWTManager) GetKeyID() string {
	if jm.useRotationKey {
		return "rotation"
	}
	return "primary"
}

func (jm *JWTManager) GenerateAccessToken(userID uint, email string, roleID *uint, role string) (string, error) {
	keyID := jm.GetKeyID()

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleID:    roleID,
		Role:      role,
		TokenType: "access",
		KeyID:     keyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(jm.expireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID
	return token.SignedString(jm.GetCurrentPrivateKey())
}

func (jm *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("metode signing tidak valid")
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return jm.publicKey, nil
		}

		if claims.KeyID == "rotation" {
			return jm.publicKeyRotation, nil
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.TokenType != "access" {
			return nil, errors.New("tipe token tidak valid")
		}
		return claims, nil
	}

	return nil, errors.New("token tidak valid")
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("gagal decode PEM block")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	}

	privateKey, ok := privateKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("kunci privat bukan RSA key")
	}

	return privateKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("gagal decode PEM block")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey = pubKeyInterface.(*rsa.PublicKey)
	}

	return publicKey, nil
}

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

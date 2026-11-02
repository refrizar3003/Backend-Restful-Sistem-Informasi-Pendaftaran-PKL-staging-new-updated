package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtService adalah antarmuka untuk operasi JWT.
type JwtService interface {
	GenerateLoginToken(userID uint, role string) (string, error)
	GenerateResetToken(userID uint) (string, error)
	GenerateVerificationToken(email string) (string, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
	GenerateLoginChallengeToken(userID uint) (string, error)
}

type jwtService struct {
	secretKey 		   string
	loginExpireHours   time.Duration
	resetExpireMinutes time.Duration
	verifyExpireHours  time.Duration
	challengeExpireMinutes time.Duration
}

func NewJwtService() JwtService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("Warning: JWT_SECRET is not set!")
		secret = "defaultsecret"
	}
	return &jwtService{
		secretKey: 			secret,
		loginExpireHours:   getEnvDuration("JWT_LOGIN_EXPIRE_HOURS", 72),
		resetExpireMinutes: getEnvDuration("JWT_RESET_EXPIRE_MINUTES", 10),
		verifyExpireHours:  getEnvDuration("JWT_VERIFY_EXPIRE_HOURS", 24),
		challengeExpireMinutes: getEnvDuration("JWT_LOGIN_CHALLENGE_EXPIRE_MINUTES", 10),
	}
}

func (s *jwtService) GenerateLoginToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"scope":   "login",
		"exp":     time.Now().Add(s.loginExpireHours * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	return s.sign(claims)
}

func (s *jwtService) GenerateResetToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"scope":   "reset_password",
		"exp":     time.Now().Add(s.resetExpireMinutes * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	return s.sign(claims)
}

func (s *jwtService) GenerateVerificationToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"scope": "verification",
		"exp":   time.Now().Add(s.verifyExpireHours * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	return s.sign(claims)
}

func (s *jwtService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}

func (s *jwtService) sign(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("Warning: invalid value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return time.Duration(parsed)
}

func (s *jwtService) GenerateLoginChallengeToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"scope":   "login_otp",
		"exp":     time.Now().Add(s.challengeExpireMinutes * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	return s.sign(claims)
}
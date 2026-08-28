package session

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	InstansiID string `json:"instansi_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

func CreateToken(userID, username, instansiID, role, secret string) (string, error) {
	claims := Claims{
		UserID:     userID,
		Username:   username,
		InstansiID: instansiID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func CreateSession(db *gorm.DB, userID string) (string, error) {
	token := uuid.New().String()
	session := models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}
	err := db.Create(&session).Error
	return token, err
}

func GetSessionUser(db *gorm.DB, token string) (*models.User, error) {
	var session models.Session
	if err := db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error; err != nil {
		return nil, err
	}

	var user models.User
	if err := db.First(&user, "id = ?", session.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteSession(db *gorm.DB, token string) error {
	return db.Where("token = ?", token).Delete(&models.Session{}).Error
}

package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"domain-manager/internal/models"
	"gorm.io/gorm"
)

// RefreshTokenService 刷新令牌服务
type RefreshTokenService struct {
	db *gorm.DB
}

// RefreshTokenModel 刷新令牌数据模型
type RefreshTokenModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null;size:64"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false;index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// 关联
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// NewRefreshTokenService 创建刷新令牌服务
func NewRefreshTokenService(db *gorm.DB) *RefreshTokenService {
	return &RefreshTokenService{db: db}
}

// GenerateRefreshToken 生成刷新令牌
func (s *RefreshTokenService) GenerateRefreshToken(userID uint) (string, error) {
	// 生成随机令牌
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", errors.New("生成刷新令牌失败")
	}
	
	tokenString := hex.EncodeToString(tokenBytes)
	
	// 清除该用户的旧刷新令牌（可选：保留最新的N个）
	if err := s.RevokeUserRefreshTokens(userID); err != nil {
		// 记录错误但不阻止新令牌生成
	}
	
	// 保存新的刷新令牌
	refreshToken := RefreshTokenModel{
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30天有效期
		IsRevoked: false,
	}
	
	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", errors.New("保存刷新令牌失败")
	}
	
	return tokenString, nil
}

// ValidateRefreshToken 验证刷新令牌
func (s *RefreshTokenService) ValidateRefreshToken(tokenString string) (*RefreshTokenModel, error) {
	var refreshToken RefreshTokenModel
	
	if err := s.db.Where("token = ? AND is_revoked = false AND expires_at > ?",
		tokenString, time.Now()).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("刷新令牌无效或已过期")
		}
		return nil, errors.New("验证刷新令牌失败")
	}
	
	return &refreshToken, nil
}

// RevokeRefreshToken 撤销指定的刷新令牌
func (s *RefreshTokenService) RevokeRefreshToken(tokenString string) error {
	result := s.db.Model(&RefreshTokenModel{}).
		Where("token = ?", tokenString).
		Update("is_revoked", true)
	
	if result.Error != nil {
		return errors.New("撤销刷新令牌失败")
	}
	
	if result.RowsAffected == 0 {
		return errors.New("刷新令牌不存在")
	}
	
	return nil
}

// RevokeUserRefreshTokens 撤销用户的所有刷新令牌
func (s *RefreshTokenService) RevokeUserRefreshTokens(userID uint) error {
	if err := s.db.Model(&RefreshTokenModel{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error; err != nil {
		return errors.New("撤销用户刷新令牌失败")
	}
	
	return nil
}

// CleanupExpiredTokens 清理过期的刷新令牌
func (s *RefreshTokenService) CleanupExpiredTokens() error {
	// 删除过期超过7天的令牌
	cutoffTime := time.Now().AddDate(0, 0, -7)
	
	if err := s.db.Where("expires_at < ? OR (is_revoked = true AND updated_at < ?)",
		time.Now(), cutoffTime).Delete(&RefreshTokenModel{}).Error; err != nil {
		return errors.New("清理过期令牌失败")
	}
	
	return nil
}

// GetUserActiveTokensCount 获取用户活跃令牌数量
func (s *RefreshTokenService) GetUserActiveTokensCount(userID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = false AND expires_at > ?",
			userID, time.Now()).Count(&count).Error; err != nil {
		return 0, errors.New("获取令牌数量失败")
	}
	
	return count, nil
}

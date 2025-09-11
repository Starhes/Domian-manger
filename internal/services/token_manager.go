package services

import (
	"fmt"
	"sync"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

// TokenManager JWT令牌管理器
type TokenManager struct {
	blacklist map[string]time.Time // token -> 过期时间
	mu        sync.RWMutex         // 读写锁保护并发访问
}

// NewTokenManager 创建新的令牌管理器
func NewTokenManager() *TokenManager {
	tm := &TokenManager{
		blacklist: make(map[string]time.Time),
	}
	
	// 启动清理goroutine
	go tm.cleanupExpiredTokens()
	
	return tm
}

// RevokeToken 撤销令牌（添加到黑名单）
func (tm *TokenManager) RevokeToken(tokenString string) error {
	// 解析token获取过期时间
	token, err := jwt.Parse(tokenString, nil)
	if err != nil {
		// 即使解析失败，也将token加入黑名单
		// 设置一个默认的过期时间（7天后）
		tm.mu.Lock()
		tm.blacklist[tokenString] = time.Now().Add(7 * 24 * time.Hour)
		tm.mu.Unlock()
		return nil
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// 无法获取claims，使用默认过期时间
		tm.mu.Lock()
		tm.blacklist[tokenString] = time.Now().Add(7 * 24 * time.Hour)
		tm.mu.Unlock()
		return nil
	}
	
	// 获取token的过期时间
	var expTime time.Time
	if exp, exists := claims["exp"]; exists {
		if expFloat, ok := exp.(float64); ok {
			expTime = time.Unix(int64(expFloat), 0)
		} else {
			expTime = time.Now().Add(7 * 24 * time.Hour)
		}
	} else {
		expTime = time.Now().Add(7 * 24 * time.Hour)
	}
	
	tm.mu.Lock()
	tm.blacklist[tokenString] = expTime
	tm.mu.Unlock()
	
	return nil
}

// IsTokenRevoked 检查令牌是否已被撤销
func (tm *TokenManager) IsTokenRevoked(tokenString string) bool {
	tm.mu.RLock()
	expTime, exists := tm.blacklist[tokenString]
	tm.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// 如果token已过期，从黑名单中移除并返回false
	// 因为过期的token本身就是无效的
	if time.Now().After(expTime) {
		tm.mu.Lock()
		delete(tm.blacklist, tokenString)
		tm.mu.Unlock()
		return false
	}
	
	return true
}

// RevokeAllUserTokens 撤销用户的所有令牌
func (tm *TokenManager) RevokeAllUserTokens(userID uint, jwtSecret string) error {
	// 这个方法会在用户更改密码、被封禁等情况下调用
	// 由于我们无法直接获取用户的所有token，我们将记录用户ID和撤销时间
	// 在验证token时检查token的签发时间是否早于撤销时间
	
	tm.mu.Lock()
	// 使用特殊的key格式来存储用户撤销时间
	revocationKey := fmt.Sprintf("user_revocation_%d", userID)
	tm.blacklist[revocationKey] = time.Now().Add(7 * 24 * time.Hour) // 保持7天
	tm.mu.Unlock()
	
	return nil
}

// IsUserTokenRevoked 检查用户的token是否因为全局撤销而无效
func (tm *TokenManager) IsUserTokenRevoked(userID uint, tokenIssuedAt time.Time) bool {
	tm.mu.RLock()
	revocationKey := fmt.Sprintf("user_revocation_%d", userID)
	revocationTime, exists := tm.blacklist[revocationKey]
	tm.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// 如果token的签发时间早于撤销时间，则该token无效
	return tokenIssuedAt.Before(revocationTime)
}

// cleanupExpiredTokens 定期清理过期的黑名单条目
func (tm *TokenManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			tm.mu.Lock()
			now := time.Now()
			for token, expTime := range tm.blacklist {
				if now.After(expTime) {
					delete(tm.blacklist, token)
				}
			}
			tm.mu.Unlock()
		}
	}
}

// GetBlacklistSize 获取黑名单大小（用于监控和调试）
func (tm *TokenManager) GetBlacklistSize() int {
	tm.mu.RLock()
	size := len(tm.blacklist)
	tm.mu.RUnlock()
	return size
}

// ClearExpiredTokens 手动清理过期令牌
func (tm *TokenManager) ClearExpiredTokens() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	now := time.Now()
	cleared := 0
	for token, expTime := range tm.blacklist {
		if now.After(expTime) {
			delete(tm.blacklist, token)
			cleared++
		}
	}
	
	return cleared
}

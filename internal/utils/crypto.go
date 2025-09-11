package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// CryptoService 加密服务
type CryptoService struct {
	key []byte
}

// NewCryptoService 创建加密服务实例
func NewCryptoService(key string) (*CryptoService, error) {
	if len(key) != 32 {
		return nil, errors.New("AES密钥必须是32字节长度")
	}
	
	return &CryptoService{
		key: []byte(key),
	}, nil
}

// Encrypt 使用AES-GCM加密数据
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("明文不能为空")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM cipher失败: %v", err)
	}

	// 生成随机nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %v", err)
	}

	// 加密数据
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// 返回十六进制编码的结果
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 使用AES-GCM解密数据
func (c *CryptoService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", errors.New("密文不能为空")
	}

	// 解码十六进制
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("十六进制解码失败: %v", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM cipher失败: %v", err)
	}

	// 检查数据长度
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文数据长度不足")
	}

	// 提取nonce和密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

// GenerateEncryptionKey 生成32字节的随机加密密钥
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("生成加密密钥失败: %v", err)
	}
	return hex.EncodeToString(key), nil
}

// ValidateEncryptionKey 验证加密密钥的有效性
func ValidateEncryptionKey(key string) error {
	if key == "" {
		return errors.New("加密密钥不能为空")
	}
	
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return errors.New("加密密钥必须是有效的十六进制字符串")
	}
	
	if len(decoded) != 32 {
		return errors.New("加密密钥解码后必须是32字节长度")
	}
	
	return nil
}

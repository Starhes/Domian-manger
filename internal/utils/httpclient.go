package utils

import (
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultHTTPClient 创建一个配置良好的HTTP客户端
func DefaultHTTPClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: &http.Transport{
			// 连接池配置
			MaxIdleConns:          100,              // 最大空闲连接数
			MaxIdleConnsPerHost:   10,               // 每个主机的最大空闲连接数
			MaxConnsPerHost:       50,               // 每个主机的最大连接数
			IdleConnTimeout:       90 * time.Second, // 空闲连接超时时间
			
			// 握手和响应超时
			TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
			ExpectContinueTimeout: 1 * time.Second,  // Expect: 100-continue 超时
			ResponseHeaderTimeout: 10 * time.Second, // 响应头超时
			
			// 拨号配置
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // 连接超时
				KeepAlive: 30 * time.Second, // TCP Keep-Alive
			}).DialContext,
			
			// 禁用HTTP/2以获得更好的兼容性（可选）
			ForceAttemptHTTP2: true,
		},
		Timeout: timeout,
	}
}

// NewRequestWithTimeout 创建带超时的HTTP请求
func NewRequestWithTimeout(ctx context.Context, method, url string, body io.Reader, timeout time.Duration) (*http.Request, context.CancelFunc, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		return req, cancel, err
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	return req, func() {}, err
}

// IsRetryableHTTPError 判断HTTP错误是否可以重试
func IsRetryableHTTPError(err error, statusCode int) bool {
	// 网络错误总是可以重试
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection") ||
			strings.Contains(errStr, "temporary failure") ||
			strings.Contains(errStr, "i/o timeout") ||
			strings.Contains(errStr, "network is unreachable") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "connection reset") {
			return true
		}
	}
	
	// HTTP状态码重试判断
	switch statusCode {
	case 429: // Too Many Requests
		return true
	case 500, 502, 503, 504: // 5xx错误（除了501 Not Implemented）
		return true
	default:
		return false
	}
}

// ExponentialBackoff 计算指数退避延迟时间
func ExponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	if baseDelay == 0 {
		baseDelay = 1 * time.Second
	}
	if maxDelay == 0 {
		maxDelay = 30 * time.Second
	}
	
	delay := time.Duration(attempt*attempt) * baseDelay
	
	// 添加随机抖动（±10%）
	jitter := time.Duration(rand.Intn(int(delay.Nanoseconds()/10))) * time.Nanosecond
	if rand.Intn(2) == 0 {
		delay += jitter
	} else {
		delay -= jitter
	}
	
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}
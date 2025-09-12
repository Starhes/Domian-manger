package utils

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int           // 失败阈值
	successThreshold int           // 成功阈值（半开状态下）
	timeout          time.Duration // 熔断器打开后的超时时间
	lastFailureTime  time.Time
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(failureThreshold int, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

var (
	ErrCircuitBreakerOpen = errors.New("熔断器已打开")
)

// Call 执行受熔断器保护的调用
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	
	// 检查熔断器状态
	switch cb.state {
	case StateOpen:
		// 检查是否可以转换为半开状态
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			cb.mu.Unlock()
			return ErrCircuitBreakerOpen
		}
	case StateHalfOpen:
		// 半开状态，允许有限的请求通过
	case StateClosed:
		// 关闭状态，正常执行
	}
	
	cb.mu.Unlock()
	
	// 执行函数
	err := fn()
	
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	
	return err
}

// onFailure 处理失败情况
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
	}
}

// onSuccess 处理成功情况
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// GetCounts 获取计数信息
func (cb *CircuitBreaker) GetCounts() (failureCount int, successCount int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureCount, cb.successCount
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// String 返回状态的字符串表示
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(s))
	}
}
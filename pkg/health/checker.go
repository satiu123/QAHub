package health

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

// HealthAware 定义了一个可以被注入健康更新器的对象的接口
type HealthAware interface {
	SetHealthUpdater(updater StatusUpdater, serviceName string)
}

// Checker 是一个可复用的健康检查组件
type Checker struct {
	updater     StatusUpdater
	serviceName string
	status      atomic.Int32
}

// NewChecker 创建一个新的健康检查器实例
func NewChecker(updater StatusUpdater, serviceName string) *Checker {
	c := &Checker{
		updater:     updater,
		serviceName: serviceName,
	}
	// 初始状态为 UNKNOWN
	c.status.Store(int32(healthv1.HealthCheckResponse_UNKNOWN))
	return c
}

// CheckAndSetStatus 执行一次健康检查，并且只在状态改变时更新
// 它接收一个 checkFunc 函数作为参数，这个函数是实际执行检查的逻辑 (e.g., db.Ping)
func (c *Checker) CheckAndSetStatus(checkFunc func(ctx context.Context) error, checkName string) {
	var newStatus healthv1.HealthCheckResponse_ServingStatus

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := checkFunc(ctx); err != nil {
		newStatus = healthv1.HealthCheckResponse_NOT_SERVING
	} else {
		newStatus = healthv1.HealthCheckResponse_SERVING
	}

	if newStatus != healthv1.HealthCheckResponse_ServingStatus(c.status.Load()) {
		c.updater.SetServingStatus(c.serviceName, newStatus)
		c.status.Store(int32(newStatus))
		log.Printf("HEALTH CHECK (%s): Service '%s' status changed to %s", checkName, c.serviceName, newStatus.String())
	}
}

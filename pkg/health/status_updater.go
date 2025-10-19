package health

import (
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

// StatusUpdater 定义了一个可以更新服务健康状态的组件的接口
type StatusUpdater interface {
	SetServingStatus(service string, status healthv1.HealthCheckResponse_ServingStatus)
}

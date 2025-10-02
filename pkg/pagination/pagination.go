package pagination

// Paginatable 接口定义了一个可以进行分页的 gRPC 请求消息。
type Paginatable interface {
	GetPage() int32
	GetPageSize() int32
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// NormalizePageAndSize 从 gRPC 请求中提取并规范化 page 和 pageSize
func NormalizePageAndSize(req Paginatable) (page int64, pageSize int32) {
	p := req.GetPage()
	ps := req.GetPageSize()

	if p < 1 {
		p = DefaultPage
	}
	if ps < 1 {
		ps = DefaultPageSize
	}
	if ps > MaxPageSize {
		ps = MaxPageSize
	}

	return int64(p), ps
}

func CalculateOffset(page int64, pageSize int32) (limit int32, offset int64) {
	return pageSize, (page - 1) * int64(pageSize)
}

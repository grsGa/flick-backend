package server

// Server 定义服务接口
type Server interface {
	// Run 启动服务
	Run(port string) error
}

// Package wego 是一个随便取的名字
package wego

import (
	"github.com/a1270107629/studyroom/sr/pkg/ginx"
	"github.com/a1270107629/studyroom/sr/pkg/grpcx"
	"github.com/a1270107629/studyroom/sr/pkg/saramax"
)

// App 当你在 wire 里面使用这个结构体的时候，要注意不是所有的服务都需要全部字段，
// 那么在 wire 的时候就不要使用 * 了
type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []saramax.Consumer
}

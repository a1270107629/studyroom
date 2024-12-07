package startup

import "github.com/a1270107629/studyroom/sr/pkg/logger"

func InitLog() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}

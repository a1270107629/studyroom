package logger

//func example() {
//	var l Logger
//	l.Info("微信用户未注册，注册新用户, 微信信息 %v", wechatInfo)
//
//	var l1 LoggerV1
//	l1.Info("微信用户未注册，注册新用户", Field{
//		Key:   "微信信息",
//		Value: wechatInfo,
//	})
//
//	var l2 LoggerV2
//	l2.Info("微信用户未注册，注册新用户",
//		"微信信息", wechatInfo)
//}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
	With(args ...Field) LoggerV1
}

type Field struct {
	Key   string
	Value any
}

// LoggerV2 要求 args 必须是偶数，第一个是 key，第二个是 value。
type LoggerV2 interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	opt := &redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
	}
	// 这里演示读取特定的某个字段
	cmd := redis.NewClient(opt)
	return cmd
}

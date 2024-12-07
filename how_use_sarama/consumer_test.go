package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	cg, err := sarama.NewConsumerGroup(addrs,
		"test_group", cfg)
	assert.NoError(t, err)
	// 这里是测试，我们就控制消费三十秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	// 开始消费，会在这里阻塞住
	err = cg.Consume(ctx,
		[]string{"test_topic"}, &ConsumerHandler{})
	assert.NoError(t, err)
}

type ConsumerHandler struct {
}

func (c *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	// 执行一些初始化的事情
	log.Println("Handler Setup")
	// 假设要重置到 0
	var offset int64 = 0
	// 遍历所有的分区
	partitions := session.Claims()["test_topic"]
	for _, p := range partitions {
		session.ResetOffset("test_topic", p, offset, "")
	}

	return nil
}

func (c *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	// 执行一些清理工作
	log.Println("Handler Cleanup")
	return nil
}

func (c *ConsumerHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	ch := claim.Messages()
	for msg := range ch {
		log.Println(msg)
		// 标记为消费成功
		session.MarkMessage(msg, "")
	}
	return nil
}

// 这个是异步消费，批量提交的例子
func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	ch := claim.Messages()
	batchSize := 10
	for {
		var eg errgroup.Group
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 这一批次已经超时了，
				// 或者，整个 consumer 被关闭了
				// 不再尝试凑够一批了
				done = true
			case msg, ok := <-ch:
				if !ok {
					cancel()
					// channel 被关闭了
					return nil
				}
				msgs = append(msgs, msg)
				eg.Go(func() error {
					log.Println("offset", msg.Offset)
					// 标记为消费成功
					time.Sleep(time.Second * 3)
					return nil
				})
			}
		}
		err := eg.Wait()
		if err == nil {
			// 这边就要都提交了
			for _, msg := range msgs {
				session.MarkMessage(msg, "")
			}
		} else {
			// 这里可以考虑重试，也可以在具体的业务逻辑里面重试
			// 也就是 eg.Go 里面重试
		}
		cancel()
	}
}

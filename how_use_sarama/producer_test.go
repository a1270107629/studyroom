package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addrs = []string{"localhost:9094"}

func TestProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	// 发送一次，不管服务端
	cfg.Producer.RequiredAcks = sarama.NoResponse
	// 发送，并且需要写入主分区
	//cfg.Producer.RequiredAcks = sarama.WaitForLocal
	// 发送，并且需要同步到所有的 ISR 上
	//cfg.Producer.RequiredAcks = sarama.WaitForAll

	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	cfg.Producer.Partitioner = sarama.NewCustomPartitioner()
	// 这个是为了兼容 JAVA，不要用
	//cfg.Producer.Partitioner = sarama.NewReferenceHashPartitioner
	producer, err := sarama.NewSyncProducer(addrs, cfg)
	assert.NoError(t, err)
	p, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("hello,这是一条消息"),
		// 会在 producer 和 consumer 之间传递
		Headers: []sarama.RecordHeader{
			{Key: []byte("header1"), Value: []byte("header1_value")},
		},
		Metadata: map[string]any{"metadata1": "metadata_value1"},
	})
	assert.NoError(t, err)
	t.Log(p, offset)
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addrs, cfg)
	assert.NoError(t, err)
	msgCh := producer.Input()
	msgCh <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("hello,这是一条异步消息"),
		// 会在 producer 和 consumer 之间传递
		Headers: []sarama.RecordHeader{
			{Key: []byte("header1"), Value: []byte("header1_value")},
		},
		Metadata: map[string]any{"metadata1": "metadata_value1"},
	}
	// 在实践中，一般是开另外一个 goroutine 来处理结果的
	select {
	case err := <-producer.Errors():
		// 这边是出错了
		val, _ := err.Msg.Value.Encode()
		t.Log(err.Err, string(val))
	case msg := <-producer.Successes():
		// 这边是成功了
		val, _ := msg.Value.Encode()
		t.Log("成功了", string(val))
	}
}

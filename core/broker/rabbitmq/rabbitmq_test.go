package rabbitmq_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/breezedup/goserver.v3/core/broker"
	"github.com/breezedup/goserver.v3/core/broker/rabbitmq"
)

func MyHandler(e broker.Event) error {
	fmt.Println(e.Topic(), ":", e.Message())
	return nil
}

func TestDurable(t *testing.T) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		t.Skip()
	}
	rabbitmq.DefaultRabbitURL = "amqp://win88:123456@192.168.1.230:5672/win88"

	b := rabbitmq.NewBroker()
	b.Init()
	if err := b.Connect(); err != nil {
		t.Logf("cant conect to broker, skip: %v", err)
		t.Skip()
	}

	b.Subscribe("test", MyHandler, broker.Queue("queue.default"),
		broker.DisableAutoAck(),
		rabbitmq.DurableQueue())

	for i := 0; i < 100; i++ {
		b.Publish("test", &broker.Message{Body: []byte("hello")})
	}
}

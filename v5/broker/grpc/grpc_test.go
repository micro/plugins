package grpc

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go-micro.dev/v5/broker"
	"go-micro.dev/v5/registry"
)

func sub(be *testing.B, c int) {
	be.StopTimer()
	m := registry.NewMemoryRegistry()
	b := NewBroker(broker.Registry(m))
	topic := uuid.New().String()

	if err := b.Init(); err != nil {
		be.Errorf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		be.Errorf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	var subs []broker.Subscriber
	done := make(chan bool, c)

	for i := 0; i < c; i++ {
		sub, err := b.Subscribe(topic, func(p broker.Event) error {
			done <- true
			m := p.Message()

			if string(m.Body) != string(msg.Body) {
				be.Errorf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
			}

			return nil
		}, broker.Queue("shared"))
		if err != nil {
			be.Errorf("Unexpected subscribe error: %v", err)
		}
		subs = append(subs, sub)
	}

	for i := 0; i < be.N; i++ {
		be.StartTimer()
		if err := b.Publish(topic, msg); err != nil {
			be.Errorf("Unexpected publish error: %v", err)
		}
		<-done
		be.StopTimer()
	}

	for _, sub := range subs {
		sub.Unsubscribe()
	}

	if err := b.Disconnect(); err != nil {
		be.Errorf("Unexpected disconnect error: %v", err)
	}
}

func pub(be *testing.B, c int) {
	be.StopTimer()
	m := registry.NewMemoryRegistry()
	b := NewBroker(broker.Registry(m))
	topic := uuid.New().String()

	if err := b.Init(); err != nil {
		be.Errorf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		be.Errorf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	done := make(chan bool, c*4)

	sub, err := b.Subscribe(topic, func(p broker.Event) error {
		done <- true
		m := p.Message()
		if string(m.Body) != string(msg.Body) {
			be.Errorf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}
		return nil
	}, broker.Queue("shared"))
	if err != nil {
		be.Errorf("Unexpected subscribe error: %v", err)
	}

	var wg sync.WaitGroup
	ch := make(chan int, c*4)
	be.StartTimer()

	for i := 0; i < c; i++ {
		go func() {
			for range ch {
				if err := b.Publish(topic, msg); err != nil {
					be.Errorf("Unexpected publish error: %v", err)
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				wg.Done()
			}
		}()
	}

	for i := 0; i < be.N; i++ {
		wg.Add(1)
		ch <- i
	}

	wg.Wait()
	be.StopTimer()
	sub.Unsubscribe()
	close(ch)
	close(done)

	if err := b.Disconnect(); err != nil {
		be.Errorf("Unexpected disconnect error: %v", err)
	}
}

func TestBroker(t *testing.T) {
	m := registry.NewMemoryRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Errorf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Errorf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	done := make(chan bool)

	sub, err := b.Subscribe("test", func(p broker.Event) error {
		m := p.Message()

		if string(m.Body) != string(msg.Body) {
			t.Errorf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}

		close(done)
		return nil
	})
	if err != nil {
		t.Errorf("Unexpected subscribe error: %v", err)
	}

	if err := b.Publish("test", msg); err != nil {
		t.Errorf("Unexpected publish error: %v", err)
	}

	<-done
	sub.Unsubscribe()
	if err := b.Disconnect(); err != nil {
		t.Errorf("Unexpected disconnect error: %v", err)
	}
}

func TestConcurrentSubBroker(t *testing.T) {
	m := registry.NewMemoryRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Errorf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Errorf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	var subs []broker.Subscriber
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		sub, err := b.Subscribe("test", func(p broker.Event) error {
			defer wg.Done()

			m := p.Message()

			if string(m.Body) != string(msg.Body) {
				t.Errorf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
			}

			return nil
		})
		if err != nil {
			t.Errorf("Unexpected subscribe error: %v", err)
		}

		wg.Add(1)
		subs = append(subs, sub)
	}

	if err := b.Publish("test", msg); err != nil {
		t.Errorf("Unexpected publish error: %v", err)
	}

	wg.Wait()

	for _, sub := range subs {
		sub.Unsubscribe()
	}

	if err := b.Disconnect(); err != nil {
		t.Errorf("Unexpected disconnect error: %v", err)
	}
}

func TestConcurrentPubBroker(t *testing.T) {
	m := registry.NewMemoryRegistry()
	b := NewBroker(broker.Registry(m))

	if err := b.Init(); err != nil {
		t.Errorf("Unexpected init error: %v", err)
	}

	if err := b.Connect(); err != nil {
		t.Errorf("Unexpected connect error: %v", err)
	}

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	var wg sync.WaitGroup

	sub, err := b.Subscribe("test", func(p broker.Event) error {
		defer wg.Done()

		m := p.Message()

		if string(m.Body) != string(msg.Body) {
			t.Errorf("Unexpected msg %s, expected %s", string(m.Body), string(msg.Body))
		}

		return nil
	})
	if err != nil {
		t.Errorf("Unexpected subscribe error: %v", err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		if err := b.Publish("test", msg); err != nil {
			t.Errorf("Unexpected publish error: %v", err)
		}
	}

	wg.Wait()

	sub.Unsubscribe()

	if err := b.Disconnect(); err != nil {
		t.Errorf("Unexpected disconnect error: %v", err)
	}
}

func BenchmarkSub1(b *testing.B) {
	sub(b, 1)
}
func BenchmarkSub8(b *testing.B) {
	sub(b, 8)
}

func BenchmarkSub32(b *testing.B) {
	sub(b, 32)
}

func BenchmarkSub64(b *testing.B) {
	sub(b, 64)
}

func BenchmarkSub128(b *testing.B) {
	sub(b, 128)
}

func BenchmarkPub1(b *testing.B) {
	pub(b, 1)
}

func BenchmarkPub8(b *testing.B) {
	pub(b, 8)
}

func BenchmarkPub32(b *testing.B) {
	pub(b, 32)
}

func BenchmarkPub64(b *testing.B) {
	pub(b, 64)
}

func BenchmarkPub128(b *testing.B) {
	pub(b, 128)
}

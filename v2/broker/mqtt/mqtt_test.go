package mqtt

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/micro/go-micro/v2/broker"
)

func TestMQTTMock(t *testing.T) {
	c := newMockClient()

	if tk := c.Connect(); tk == nil {
		t.Error("got nil token")
	}

	if tk := c.Subscribe("mock", 0, func(cm mqtt.Client, m mqtt.Message) {
		t.Logf("Received payload %+v", string(m.Payload()))
	}); tk == nil {
		t.Error("got nil token")
	}

	if tk := c.Publish("mock", 0, false, []byte(`hello world`)); tk == nil {
		t.Error("got nil token")
	}

	if tk := c.Unsubscribe("mock"); tk == nil {
		t.Error("got nil token")
	}

	c.Disconnect(0)
}

func TestMQTTHandler(t *testing.T) {
	p := &mqttPub{
		topic: "mock",
		msg:   &broker.Message{Body: []byte(`hello`)},
	}

	if p.Topic() != "mock" {
		t.Error("Expected topic mock got", p.Topic())
	}

	if string(p.Message().Body) != "hello" {
		t.Errorf("Expected `hello` message got %s", string(p.Message().Body))
	}

	s := &mqttSub{
		topic:  "mock",
		client: newMockClient(),
	}

	s.client.Connect()

	if s.Topic() != "mock" {
		t.Error("Expected topic mock got", s.Topic())
	}

	if err := s.Unsubscribe(); err != nil {
		t.Error("Error unsubscribing", err)
	}

	s.client.Disconnect(0)
}

func TestMQTT(t *testing.T) {
	b := NewBroker()

	if err := b.Init(); err != nil {
		t.Error(err)
	}

	// use mock client
	b.(*mqttBroker).client = newMockClient()

	if tk := b.(*mqttBroker).client.Connect(); tk == nil {
		t.Error("got nil token")
	}

	if err := b.Publish("mock", &broker.Message{Body: []byte(`hello`)}); err != nil {
		t.Error(err)
	}

	if err := b.Disconnect(); err != nil {
		t.Error(err)
	}

	b.(*mqttBroker).client.Disconnect(0)
}

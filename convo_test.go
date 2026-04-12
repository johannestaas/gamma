package gamma_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/johannestaas/gamma"
)

func TestConvoNoPrompt(t *testing.T) {
	opts := []gamma.Option{
		gamma.WithHTTPClient(testClient()),
	}
	c := gamma.NewGammaClient(opts...)
	convo := c.NewConvo("")
	if len(convo.Messages) > 0 {
		t.Fatalf("messages was not empty: %+v", convo.Messages)
	}
}

func TestConvoWithPrompt(t *testing.T) {
	opts := []gamma.Option{
		gamma.WithHTTPClient(testClient()),
	}
	c := gamma.NewGammaClient(opts...)
	convo := c.NewConvo("This is a prompt")
	if len(convo.Messages) != 1 {
		t.Fatalf("messages was not len 1: %+v", convo.Messages)
	}
}

func TestConvoAsk(t *testing.T) {
	opts := []gamma.Option{
		gamma.WithHTTPClient(testClient()),
	}
	c := gamma.NewGammaClient(opts...)
	convo := c.NewConvo("This is a prompt")
	convo.Ask(context.Background(), "Hello there")
	if len(convo.Messages) != 2 {
		t.Fatalf("messages: %+v", convo.Messages)
	}
	if !reflect.DeepEqual(convo.Messages[0], gamma.Message{
		Role:    "system",
		Content: "This is a prompt",
	}) {
		t.Fatalf("messages[0] was wrong: %+v", convo.Messages[0])
	}
	if !reflect.DeepEqual(convo.Messages[1], gamma.Message{
		Role:    "user",
		Content: "Hello there",
	}) {
		t.Fatalf("messages[1] was wrong: %+v", convo.Messages[1])
	}
}

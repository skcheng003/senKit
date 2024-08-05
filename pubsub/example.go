package pubsub

import (
	"fmt"
	"strings"
	"time"
)

func Example() {
	publisher := NewPublisher(100*time.Millisecond, 10)
	defer publisher.Close()

	all := publisher.Subscribe()
	golang := publisher.SubscribeTopic(func(v any) bool {
		if s, ok := v.(string); ok {
			return strings.Contains(s, "golang")
		}
		return false
	})
	publisher.Publish("hello, I am")
	publisher.Publish("learning golang")

	go func() {
		for msg := range all {
			fmt.Println(msg)
		}
	}()

	go func() {
		for msg := range golang {
			fmt.Println(msg)
		}
	}()

	time.Sleep(time.Second * 3)
}

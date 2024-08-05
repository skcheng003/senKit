package pubsub

import (
	"sync"
	"time"
)

type (
	subscriber chan any
	topicFunc  func(v any) bool
)

// Publisher 发布者对象
type Publisher struct {
	mu          sync.RWMutex             // 读写锁
	buffer      int                      // 订阅队列的缓冲大小
	timeout     time.Duration            // 发布超时时间
	subscribers map[subscriber]topicFunc // 订阅者信息
}

func NewPublisher(timeout time.Duration, buffer int) *Publisher {
	return &Publisher{
		buffer:      buffer,
		timeout:     timeout,
		subscribers: make(map[subscriber]topicFunc),
	}
}

// Subscribe 添加一个新的订阅者，订阅全部主题
func (p *Publisher) Subscribe() chan any {
	return p.SubscribeTopic(nil)
}

// SubscribeTopic 添加一个新的订阅者，订阅过滤器筛选后的主题
func (p *Publisher) SubscribeTopic(topicFunc topicFunc) chan any {
	ch := make(chan any, p.buffer)
	p.mu.Lock()
	p.subscribers[ch] = topicFunc
	p.mu.Unlock()
	return ch
}

// Evict 退出订阅
func (p *Publisher) Evict(sub chan any) {
	p.mu.RLock()
	defer p.mu.Unlock()
	delete(p.subscribers, sub)
	close(sub)
}

// Publish 发布一个主题
func (p *Publisher) Publish(v any) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var wg sync.WaitGroup
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, v, &wg)
	}
	wg.Wait()
}

// 发送主题，可以容忍一定的超时
func (p *Publisher) sendTopic(sub chan any, topic topicFunc, v any, wg *sync.WaitGroup) {
	defer wg.Done()
	if topic != nil && !topic(v) {
		return
	}
	// select 实现管道的超时判断
	select {
	case sub <- v:
	case <-time.After(p.timeout):
	}
}

// Close 关闭发布者对象，同时关闭所有订阅者通道
func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

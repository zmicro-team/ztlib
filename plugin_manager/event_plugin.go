package pluginmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventTypeStartup  EventType = "startup"
	EventTypeShutdown EventType = "shutdown"
	EventTypeTick     EventType = "tick"
	EventTypeCustom   EventType = "custom"
)

// Event 事件结构
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"source"`
}

// EventPlugin 事件插件实现
type EventPlugin struct {
	name             string
	running          bool
	events           chan Event
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	eventSubscribers map[string]chan Event
	subMu            sync.RWMutex
	ticker           *time.Ticker
	config           EventPluginConfig
}

// EventPluginConfig 事件插件配置
type EventPluginConfig struct {
	TickInterval    time.Duration `json:"tick_interval"`     // 定时事件间隔
	EventBufferSize int           `json:"event_buffer_size"` // 事件通道缓冲大小
	EnableLogging   bool          `json:"enable_logging"`    // 是否启用日志
}

// DefaultEventPluginConfig 默认配置
func DefaultEventPluginConfig() EventPluginConfig {
	return EventPluginConfig{
		TickInterval:    time.Second * 5,
		EventBufferSize: 100,
		EnableLogging:   true,
	}
}

// NewEventPlugin 创建新的事件插件
func NewEventPlugin(name string, config EventPluginConfig) *EventPlugin {
	if config.EventBufferSize <= 0 {
		config.EventBufferSize = 100
	}
	if config.TickInterval <= 0 {
		config.TickInterval = time.Second * 5
	}

	return &EventPlugin{
		name:             name,
		events:           make(chan Event, config.EventBufferSize),
		eventSubscribers: make(map[string]chan Event),
		config:           config,
	}
}

// Start 启动事件插件
func (ep *EventPlugin) Start(ctx context.Context) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if ep.running {
		return fmt.Errorf("event plugin '%s' is already running", ep.name)
	}

	ep.ctx, ep.cancel = context.WithCancel(ctx)
	ep.running = true

	// 启动事件处理器
	go ep.eventLoop()

	// 启动定时器
	if ep.config.TickInterval > 0 {
		ep.ticker = time.NewTicker(ep.config.TickInterval)
		go ep.tickLoop()
	}

	// 发送启动事件
	ep.sendEvent(Event{
		Type:      EventTypeStartup,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "Event plugin started"},
		Source:    ep.name,
	})

	if ep.config.EnableLogging {
		log.Printf("Event plugin '%s' started", ep.name)
	}

	return nil
}

// Stop 停止事件插件
func (ep *EventPlugin) Stop(ctx context.Context) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if !ep.running {
		return fmt.Errorf("event plugin '%s' is not running", ep.name)
	}

	// 发送关闭事件
	ep.sendEvent(Event{
		Type:      EventTypeShutdown,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "Event plugin stopping"},
		Source:    ep.name,
	})

	// 取消上下文
	if ep.cancel != nil {
		ep.cancel()
	}

	// 停止定时器
	if ep.ticker != nil {
		ep.ticker.Stop()
	}

	// 关闭所有订阅者通道
	ep.subMu.Lock()
	for id, ch := range ep.eventSubscribers {
		close(ch)
		delete(ep.eventSubscribers, id)
	}
	ep.subMu.Unlock()

	// 关闭事件通道
	close(ep.events)
	ep.running = false

	if ep.config.EnableLogging {
		log.Printf("Event plugin '%s' stopped", ep.name)
	}

	return nil
}

// Events 返回事件通道
func (ep *EventPlugin) Events(ctx context.Context, value any) <-chan any {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.running {
		// 返回已关闭的通道
		ch := make(chan any)
		close(ch)
		return ch
	}

	// 直接返回 ep.events 转换为 any 通道
	resultCh := make(chan any, cap(ep.events))
	go func() {
		defer close(resultCh)
		for {
			select {
			case event, ok := <-ep.events:
				if !ok {
					return
				}
				select {
				case resultCh <- event:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultCh
}

// SyncEvents 同步获取事件
func (ep *EventPlugin) SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error) {
	if !ep.IsRunning() {
		return nil, fmt.Errorf("plugin '%s' is not running", ep.name)
	}

	// 如果指定了超时，创建带超时的 context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case event := <-ep.events:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Name 返回插件名称
func (ep *EventPlugin) Name() string {
	return ep.name
}

// IsRunning 检查插件是否正在运行
func (ep *EventPlugin) IsRunning() bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.running
}

// PublishEvent 发布自定义事件
func (ep *EventPlugin) PublishEvent(eventType EventType, data map[string]interface{}) error {
	if !ep.IsRunning() {
		return fmt.Errorf("plugin '%s' is not running", ep.name)
	}

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Source:    ep.name,
	}

	ep.sendEvent(event)
	return nil
}

// SubscribeToEvents 订阅特定类型的事件
func (ep *EventPlugin) SubscribeToEvents(eventType EventType) (<-chan Event, func()) {
	if !ep.IsRunning() {
		ch := make(chan Event)
		close(ch)
		return ch, func() {}
	}

	ep.subMu.Lock()
	defer ep.subMu.Unlock()

	subscriberID := fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano())
	ch := make(chan Event, 10) // 小缓冲区用于订阅

	// 将订阅信息存储，包含通道和事件类型
	ep.eventSubscribers[subscriberID] = ch

	unsubscribe := func() {
		ep.subMu.Lock()
		defer ep.subMu.Unlock()
		if subCh, exists := ep.eventSubscribers[subscriberID]; exists {
			close(subCh)
			delete(ep.eventSubscribers, subscriberID)
		}
	}

	return ch, unsubscribe
}

// GetConfig 获取插件配置
func (ep *EventPlugin) GetConfig() EventPluginConfig {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.config
}

// UpdateConfig 更新插件配置
func (ep *EventPlugin) UpdateConfig(config EventPluginConfig) error {
	if !ep.IsRunning() {
		ep.config = config
		return nil
	}

	// 如果正在运行且定时器间隔改变，需要重启定时器
	if config.TickInterval != ep.config.TickInterval && config.TickInterval > 0 {
		if ep.ticker != nil {
			ep.ticker.Stop()
		}
		ep.ticker = time.NewTicker(config.TickInterval)
	}

	ep.mu.Lock()
	ep.config = config
	ep.mu.Unlock()

	return nil
}

// 内部方法：发送事件
func (ep *EventPlugin) sendEvent(event Event) {
	select {
	case ep.events <- event:
		// 事件发送成功，通知订阅者
		ep.notifySubscribers(event)
		if ep.config.EnableLogging {
			log.Printf("Event published: %+v", event)
		}
	default:
		// 事件通道已满，丢弃事件
		if ep.config.EnableLogging {
			log.Printf("Event dropped: %+v (channel full)", event)
		}
	}
}

// 内部方法：通知订阅者
func (ep *EventPlugin) notifySubscribers(event Event) {
	ep.subMu.RLock()
	defer ep.subMu.RUnlock()

	for subscriberID, subscriberCh := range ep.eventSubscribers {
		// 从订阅ID中提取事件类型
		var subscribedType EventType
		if idx := findIndex(subscriberID, "-"); idx > 0 {
			subscribedType = EventType(subscriberID[:idx])
		}

		// 只向订阅了匹配事件类型的订阅者发送事件
		if subscribedType == "" || subscribedType == event.Type {
			select {
			case subscriberCh <- event:
			default:
				// 订阅者通道已满，记录并跳过
				if ep.config.EnableLogging {
					log.Printf("Subscriber %s channel full, event dropped", subscriberID)
				}
			}
		}
	}
}

// 辅助函数：查找字符串中的分隔符
func findIndex(s string, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// 内部方法：事件循环
func (ep *EventPlugin) eventLoop() {
	for {
		select {
		case <-ep.ctx.Done():
			return
		}
	}
}

// 内部方法：定时事件循环
func (ep *EventPlugin) tickLoop() {
	if ep.ticker == nil {
		return
	}

	for {
		select {
		case <-ep.ticker.C:
			ep.sendEvent(Event{
				Type:      EventTypeTick,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"tick": time.Now().Unix()},
				Source:    ep.name,
			})
		case <-ep.ctx.Done():
			return
		}
	}
}

package pluginmanager

import (
	"context"
	"log"
	"time"
)

// ExampleUsage 演示事件插件的使用方法
func ExampleUsage() {
	// 创建插件管理器
	pm := NewPluginManager()

	// 创建事件插件配置
	config := DefaultEventPluginConfig()
	config.TickInterval = time.Second * 2
	config.EnableLogging = true

	// 创建事件插件
	eventPlugin := NewEventPlugin("my-event-plugin", config)

	// 注册插件
	err := pm.RegisterPlugin(eventPlugin)
	if err != nil {
		log.Fatalf("Failed to register event plugin: %v", err)
	}

	// 启动所有插件
	ctx := context.Background()
	err = pm.StartAll(ctx)
	if err != nil {
		log.Fatalf("Failed to start plugins: %v", err)
	}
	defer pm.StopAll()

	// 订阅特定类型的事件
	eventCh, unsubscribe := eventPlugin.SubscribeToEvents(EventTypeTick)
	defer unsubscribe()

	// 发布自定义事件
	err = eventPlugin.PublishEvent(EventTypeCustom, map[string]interface{}{
		"message": "Hello from custom event!",
		"user_id": 12345,
	})
	if err != nil {
		log.Printf("Failed to publish custom event: %v", err)
	}

	// 监听事件
	go func() {
		for event := range eventCh {
			log.Printf("Received tick event: %+v at %v", event.Data, event.Timestamp)
		}
	}()

	// 同步获取事件
	go func() {
		for i := 0; i < 3; i++ {
			event, err := pm.GetSyncEvent("my-event-plugin", "request", time.Second*3)
			if err != nil {
				log.Printf("Failed to get sync event: %v", err)
				continue
			}

			if e, ok := event.(Event); ok {
				log.Printf("Sync event received: Type=%s, Data=%v", e.Type, e.Data)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	// 运行一段时间
	time.Sleep(time.Second * 10)

	// 获取所有插件的同步事件
	events, err := pm.GetAllSyncEvents("final_request", time.Second*2)
	if err != nil {
		log.Printf("Failed to get all sync events: %v", err)
	} else {
		log.Printf("All sync events: %+v", events)
	}
}

// AdvancedExample 高级用法示例
func AdvancedExample() {
	// 创建多个事件插件
	pm := NewPluginManager()

	// 系统事件插件
	systemPlugin := NewEventPlugin("system-events", DefaultEventPluginConfig())

	// 用户事件插件
	userConfig := DefaultEventPluginConfig()
	userConfig.TickInterval = time.Second * 5
	userPlugin := NewEventPlugin("user-events", userConfig)

	// 注册插件
	pm.RegisterPlugin(systemPlugin)
	pm.RegisterPlugin(userPlugin)

	// 启动插件
	ctx := context.Background()
	pm.StartAll(ctx)
	defer pm.StopAll()

	// 从不同插件发布事件
	systemPlugin.PublishEvent(EventTypeCustom, map[string]interface{}{
		"system_status": "healthy",
		"cpu_usage":     45.2,
	})

	userPlugin.PublishEvent(EventTypeCustom, map[string]interface{}{
		"user_action": "login",
		"user_id":     "user123",
		"timestamp":   time.Now(),
	})

	// 批量获取所有插件的事件
	events, err := pm.GetSyncEvents([]string{"system-events", "user-events"}, "batch_request", time.Second*2)
	if err != nil {
		log.Printf("Batch sync events error: %v", err)
		return
	}

	for pluginName, event := range events {
		if e, ok := event.(Event); ok {
			log.Printf("Plugin %s sent event: Type=%s, Data=%v", pluginName, e.Type, e.Data)
		}
	}
}
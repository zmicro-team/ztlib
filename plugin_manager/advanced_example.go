package pluginmanager

import (
	"context"
	"fmt"
	"log"
	"time"
)

// AdvancedExampleUsage 演示高级事件插件的使用
func AdvancedExampleUsage() {
	// 创建高级事件插件
	config := DefaultAdvancedEventPluginConfig()
	config.EventBufferSize = 100
	config.EnableStats = true
	config.StatsInterval = time.Second * 10
	config.EnableLogging = true

	plugin := NewAdvancedEventPlugin("advanced-event-plugin", config)

	// 添加过滤器
	// 只允许 startup 和 custom 类型的事件
	eventTypeFilter := NewEventTypeFilter(EventTypeStartup, EventTypeCustom, EventTypeTick)
	plugin.AddFilter(eventTypeFilter)

	// 添加数据过滤器：只允许包含重要标志的事件
	importantFilter := NewDataFilter("important", "eq", true)
	plugin.AddFilter(importantFilter)

	// 添加路由处理器
	// 高优先级：处理重要事件
	plugin.AddRoute("important-handler", func(event Event) error {
		log.Printf("IMPORTANT: %+v", event)
		return nil
	}, 1000)

	// 中优先级：处理自定义事件
	plugin.AddRoute("custom-handler", func(event Event) error {
		log.Printf("Custom Event: %s - %v", event.Type, event.Data)
		return nil
	}, 500)

	// 低优先级：统计处理器
	plugin.AddRoute("stats-handler", func(event Event) error {
		log.Printf("Stats: Event type %s processed", event.Type)
		return nil
	}, 100)

	// 启动插件
	ctx := context.Background()
	err := plugin.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start advanced event plugin: %v", err)
	}
	defer plugin.Stop(ctx)

	// 发布各种事件
	// 这个事件会被过滤掉（不是允许的类型）
	plugin.PublishEvent(Event{
		Type:      EventTypeShutdown,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "shutdown"},
		Source:    "example",
	})

	// 这个事件会被数据过滤器过滤掉
	plugin.PublishEvent(Event{
		Type:      EventTypeCustom,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"important": false, "message": "not important"},
		Source:    "example",
	})

	// 这个事件会通过所有过滤器并被路由处理
	plugin.PublishEvent(Event{
		Type:      EventTypeCustom,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{
			"important": true,
			"message":   "This is important!",
			"user_id":   12345,
		},
		Source: "example",
	})

	// 启动事件会通过类型过滤器但被数据过滤器过滤
	plugin.PublishEvent(Event{
		Type:      EventTypeStartup,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "system startup"},
		Source:    "example",
	})

	// 等待处理
	time.Sleep(time.Second * 2)

	// 查看统计信息
	stats := plugin.GetStats()
	fmt.Printf("Event Statistics:\n")
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Filtered Events: %d\n", stats.FilteredEvents)
	fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
	fmt.Printf("  Failed Routes: %d\n", stats.FailedRoutes)
	fmt.Printf("  Event Types: %v\n", stats.EventTypes)
}

// FilterChainExample 演示过滤器链的使用
func FilterChainExample() {
	config := DefaultAdvancedEventPluginConfig()
	config.EnableLogging = true
	plugin := NewAdvancedEventPlugin("filter-chain-plugin", config)

	// 创建多个过滤器
	// 1. 类型过滤器：只允许 custom 类型
	typeFilter := NewEventTypeFilter(EventTypeCustom)
	plugin.AddFilter(typeFilter)

	// 2. 数据过滤器：只允许包含 user_id 的事件
	userFilter := NewDataFilter("user_id", "ne", nil)
	plugin.AddFilter(userFilter)

	// 3. 数据过滤器：只允许活跃用户
	activeFilter := NewDataFilter("active", "eq", true)
	plugin.AddFilter(activeFilter)

	// 添加路由
	plugin.AddRoute("user-action-handler", func(event Event) error {
		userID := event.Data["user_id"]
		action := event.Data["action"]
		log.Printf("User Action: User=%v, Action=%v", userID, action)
		return nil
	}, 100)

	// 启动插件
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	// 测试不同的事件
	testEvents := []Event{
		// 会被类型过滤器过滤
		{
			Type:      EventTypeTick,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"user_id": 1, "action": "login", "active": true},
			Source:    "test",
		},
		// 会被用户ID过滤器过滤
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"action": "system", "active": true},
			Source:    "test",
		},
		// 会被活跃状态过滤器过滤
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"user_id": 2, "action": "login", "active": false},
			Source:    "test",
		},
		// 会通过所有过滤器
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"user_id": 3, "action": "logout", "active": true},
			Source:    "test",
		},
	}

	for _, event := range testEvents {
		plugin.PublishEvent(event)
		time.Sleep(time.Millisecond * 100)
	}

	time.Sleep(time.Second)

	// 显示统计
	stats := plugin.GetStats()
	fmt.Printf("Filter Chain Statistics:\n")
	fmt.Printf("  Published: %d\n", len(testEvents))
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Filtered Events: %d\n", stats.FilteredEvents)
	fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
}

// PriorityRoutingExample 演示优先级路由
func PriorityRoutingExample() {
	config := DefaultAdvancedEventPluginConfig()
	config.EnableLogging = true
	plugin := NewAdvancedEventPlugin("priority-routing-plugin", config)

	// 添加不同优先级的路由处理器
	// 最高优先级：错误处理器
	plugin.AddRoute("error-handler", func(event Event) error {
		log.Printf("[ERROR] Critical event: %+v", event)
		return nil
	}, 1000)

	// 高优先级：安全处理器
	plugin.AddRoute("security-handler", func(event Event) error {
		log.Printf("[SECURITY] Security event: %+v", event)
		return nil
	}, 800)

	// 中优先级：业务逻辑处理器
	plugin.AddRoute("business-handler", func(event Event) error {
		log.Printf("[BUSINESS] Business event: %+v", event)
		return nil
	}, 500)

	// 低优先级：日志处理器
	plugin.AddRoute("logging-handler", func(event Event) error {
		log.Printf("[LOG] Event logged: Type=%s, Source=%s", event.Type, event.Source)
		return nil
	}, 100)

	// 启动插件
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	// 发布不同优先级的事件
	events := []Event{
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"priority": "low", "message": "Regular business event"},
			Source:    "business",
		},
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"priority": "security", "message": "Security alert"},
			Source:    "security",
		},
		{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"priority": "error", "message": "Critical error occurred"},
			Source:    "error",
		},
	}

	for _, event := range events {
		plugin.PublishEvent(event)
		time.Sleep(time.Millisecond * 200)
	}

	time.Sleep(time.Second)
	stats := plugin.GetStats()
	fmt.Printf("Priority Routing Statistics:\n")
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
	fmt.Printf("  Failed Routes: %d\n", stats.FailedRoutes)
}

// DynamicConfigurationExample 演示动态配置
func DynamicConfigurationExample() {
	config := DefaultAdvancedEventPluginConfig()
	config.EnableStats = true
	config.EnableLogging = true
	plugin := NewAdvancedEventPlugin("dynamic-config-plugin", config)

	// 启动插件
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	// 添加初始路由
	plugin.AddRoute("initial-route", func(event Event) error {
		log.Printf("Initial route handler: %+v", event)
		return nil
	}, 100)

	// 发布一些事件
	for i := 0; i < 3; i++ {
		plugin.PublishEvent(Event{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i, "phase": "initial"},
			Source:    "example",
		})
		time.Sleep(time.Millisecond * 200)
	}

	// 动态添加新的高优先级路由
	plugin.AddRoute("high-priority-route", func(event Event) error {
		log.Printf("High priority handler: %+v", event)
		return nil
	}, 1000)

	// 动态添加过滤器
	filter := NewDataFilter("phase", "eq", "filtered")
	plugin.AddFilter(filter)

	// 发布更多事件
	for i := 3; i < 6; i++ {
		plugin.PublishEvent(Event{
			Type:      EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i, "phase": "filtered"},
			Source:    "example",
		})
		time.Sleep(time.Millisecond * 200)
	}

	time.Sleep(time.Second)

	// 显示最终统计
	stats := plugin.GetStats()
	fmt.Printf("Dynamic Configuration Final Stats:\n")
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Filtered Events: %d\n", stats.FilteredEvents)
	fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
	fmt.Printf("  Failed Routes: %d\n", stats.FailedRoutes)

	// 重置统计并继续
	plugin.ResetStats()
	
	// 移除过滤器
	removed := plugin.RemoveFilter(filter.Description())
	fmt.Printf("Removed %d filters\n", removed)

	// 移除路由
	routeRemoved := plugin.RemoveRoute("initial-route")
	fmt.Printf("Route removed: %v\n", routeRemoved)

	// 发布最终事件
	plugin.PublishEvent(Event{
		Type:      EventTypeCustom,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index": 6, "phase": "final"},
		Source:    "example",
	})

	time.Sleep(time.Second * 2)

	// 显示重置后的统计
	finalStats := plugin.GetStats()
	fmt.Printf("After Reset Stats:\n")
	fmt.Printf("  Total Events: %d\n", finalStats.TotalEvents)
	fmt.Printf("  Filtered Events: %d\n", finalStats.FilteredEvents)
	fmt.Printf("  Routed Events: %d\n", finalStats.RoutedEvents)
}
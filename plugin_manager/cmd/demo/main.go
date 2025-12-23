package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zmicro-team/ztlib/plugin_manager"
)

func main() {
	fmt.Println("=== Event Plugin Demo ===")
	
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 创建信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// 运行演示
	runDemo(ctx)
}

func runDemo(ctx context.Context) {
	fmt.Println("\n1. Basic Event Plugin Demo:")
	basicDemo(ctx)

	time.Sleep(time.Second * 2)

	fmt.Println("\n2. Advanced Event Plugin Demo:")
	advancedDemo(ctx)

	time.Sleep(time.Second * 2)

	fmt.Println("\n3. Integration Demo:")
	integrationDemo(ctx)
}

func basicDemo(ctx context.Context) {
	// 创建插件管理器
	pm := pluginmanager.NewPluginManager()

	// 创建基础事件插件
	config := pluginmanager.DefaultEventPluginConfig()
	config.TickInterval = time.Second
	config.EnableLogging = true

	eventPlugin := pluginmanager.NewEventPlugin("basic-plugin", config)

	// 注册插件
	err := pm.RegisterPlugin(eventPlugin)
	if err != nil {
		log.Printf("Failed to register basic plugin: %v", err)
		return
	}

	// 启动管理器
	err = pm.StartAll(ctx)
	if err != nil {
		log.Printf("Failed to start plugin manager: %v", err)
		return
	}

	// 发布一些测试事件
	eventPlugin.PublishEvent(pluginmanager.EventTypeCustom, map[string]interface{}{
		"message": "Basic plugin started",
		"demo":    "basic",
	})

	// 监听几个事件
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			event, err := pm.GetSyncEvent("basic-plugin", "test", time.Second*2)
			if err != nil {
				log.Printf("Failed to get event: %v", err)
			} else {
				if e, ok := event.(pluginmanager.Event); ok {
					log.Printf("Basic Event: Type=%s, Data=%v", e.Type, e.Data)
				}
			}
			time.Sleep(time.Second)
		}
	}

	// 停止插件
	err = pm.StopAll()
	if err != nil {
		log.Printf("Error stopping basic plugin: %v", err)
	}
}

func advancedDemo(ctx context.Context) {
	// 创建高级事件插件
	config := pluginmanager.DefaultAdvancedEventPluginConfig()
	config.EventBufferSize = 50
	config.EnableStats = true
	config.StatsInterval = time.Second * 2
	config.EnableLogging = false

	plugin := pluginmanager.NewAdvancedEventPlugin("advanced-plugin", config)

	// 添加过滤器
	eventTypeFilter := pluginmanager.NewEventTypeFilter(
		pluginmanager.EventTypeCustom,
		pluginmanager.EventTypeTick,
	)
	plugin.AddFilter(eventTypeFilter)

	// 添加路由处理器
	plugin.AddRoute("demo-handler", func(event pluginmanager.Event) error {
		log.Printf("Advanced Event Handler: %s - %+v", event.Type, event.Data)
		return nil
	}, 100)

	// 启动插件
	err := plugin.Start(ctx)
	if err != nil {
		log.Printf("Failed to start advanced plugin: %v", err)
		return
	}
	defer plugin.Stop(ctx)

	// 发布测试事件
	testEvents := []pluginmanager.Event{
		{
			Type:      pluginmanager.EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"demo": "advanced", "message": "Test message 1"},
			Source:    "demo",
		},
		{
			Type:      pluginmanager.EventTypeCustom,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"demo": "advanced", "message": "Test message 2"},
			Source:    "demo",
		},
	}

	for _, event := range testEvents {
		err := plugin.PublishEvent(event)
		if err != nil {
			log.Printf("Failed to publish event: %v", err)
		}
		time.Sleep(time.Millisecond * 500)
	}

	// 等待处理
	time.Sleep(time.Second * 2)

	// 显示统计信息
	stats := plugin.GetStats()
	fmt.Printf("Advanced Plugin Stats:\n")
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
	fmt.Printf("  Event Types: %v\n", stats.EventTypes)
}

func integrationDemo(ctx context.Context) {
	// 创建插件管理器
	pm := pluginmanager.NewPluginManager()

	// 创建多个事件插件
	basicConfig := pluginmanager.DefaultEventPluginConfig()
	basicConfig.TickInterval = time.Second * 2
	basicConfig.EnableLogging = false
	
	basicPlugin1 := pluginmanager.NewEventPlugin("system-events", basicConfig)
	basicPlugin2 := pluginmanager.NewEventPlugin("user-events", basicConfig)

	// 创建高级插件
	advancedConfig := pluginmanager.DefaultAdvancedEventPluginConfig()
	advancedConfig.EnableStats = true
	advancedConfig.EnableLogging = false
	
	advancedPlugin := pluginmanager.NewAdvancedEventPlugin("router-plugin", advancedConfig)

	// 配置高级插件的过滤器
	advancedPlugin.AddFilter(pluginmanager.NewEventTypeFilter(pluginmanager.EventTypeCustom))

	// 配置高级插件的路由
	advancedPlugin.AddRoute("integration-handler", func(event pluginmanager.Event) error {
		log.Printf("Integration Router: Received event from %s: %+v", event.Source, event.Data)
		return nil
	}, 100)

	// 注册所有插件
	plugins := []pluginmanager.CollectorPlugin{
		basicPlugin1,
		basicPlugin2,
		advancedPlugin,
	}

	for _, p := range plugins {
		err := pm.RegisterPlugin(p)
		if err != nil {
			log.Printf("Failed to register plugin %s: %v", p.Name(), err)
			continue
		}
	}

	// 启动管理器
	err := pm.StartAll(ctx)
	if err != nil {
		log.Printf("Failed to start plugin manager: %v", err)
		return
	}
	defer pm.StopAll()

	// 从不同插件发布事件
	basicPlugin1.PublishEvent(pluginmanager.EventTypeCustom, map[string]interface{}{
		"source": "system",
		"event":  "cpu_usage",
		"value":  75.5,
	})

	basicPlugin2.PublishEvent(pluginmanager.EventTypeCustom, map[string]interface{}{
		"source": "user",
		"event":  "login",
		"user_id": "user123",
	})

	advancedPlugin.PublishEvent(pluginmanager.Event{
		Type:      pluginmanager.EventTypeCustom,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"source": "router", "message": "Router is working"},
		Source:    "router-plugin",
	})

	// 等待事件处理
	time.Sleep(time.Second * 3)

	// 获取所有插件的同步事件
	fmt.Println("Getting sync events from all plugins...")
	allEvents, err := pm.GetAllSyncEvents("integration-test", time.Second*2)
	if err != nil {
		log.Printf("Failed to get all sync events: %v", err)
	} else {
		for pluginName, event := range allEvents {
			if e, ok := event.(pluginmanager.Event); ok {
				log.Printf("Event from %s: Type=%s, Source=%s", 
					pluginName, e.Type, e.Source)
			}
		}
	}

	// 显示插件状态
	fmt.Println("\nPlugin Status:")
	for _, name := range pm.GetPlugins() {
		if plugin, exists := pm.GetPlugin(name); exists {
			fmt.Printf("  %s: Running=%v\n", name, plugin.IsRunning())
		}
	}
}
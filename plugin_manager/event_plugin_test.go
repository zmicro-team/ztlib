package pluginmanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventPlugin_Lifecycle 测试事件插件的生命周期
func TestEventPlugin_Lifecycle(t *testing.T) {
	config := DefaultEventPluginConfig()
	config.TickInterval = time.Millisecond * 100 // 快速测试
	plugin := NewEventPlugin("test-event-plugin", config)

	ctx := context.Background()

	// 测试初始状态
	assert.Equal(t, "test-event-plugin", plugin.Name())
	assert.False(t, plugin.IsRunning())

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	assert.True(t, plugin.IsRunning())

	// 停止插件
	err = plugin.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, plugin.IsRunning())

	// 重复停止应该返回错误
	err = plugin.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// TestEventPlugin_Events 测试事件通道功能
func TestEventPlugin_Events(t *testing.T) {
	config := DefaultEventPluginConfig()
	config.EventBufferSize = 10
	plugin := NewEventPlugin("events-test-plugin", config)

	ctx := context.Background()

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 发布自定义事件
	err = plugin.PublishEvent(EventTypeCustom, map[string]interface{}{
		"test": "data",
		"id":   123,
	})
	require.NoError(t, err)

	// 获取事件通道
	eventCh := plugin.Events(ctx, nil)

	// 接收事件 - 会先收到startup，然后是custom
	var customEvent Event
	eventCount := 0
	
	for eventCount < 2 { // 期望收到2个事件：startup + custom
		select {
		case event := <-eventCh:
			if e, ok := event.(Event); ok {
				if e.Type == EventTypeCustom {
					customEvent = e
				}
				eventCount++
			} else {
				t.Errorf("Expected Event type, got %T", event)
				return
			}
		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for event, received %d events", eventCount)
			return
		}
	}

	// 验证custom事件
	assert.Equal(t, EventTypeCustom, customEvent.Type)
	assert.Equal(t, "events-test-plugin", customEvent.Source)
	assert.Equal(t, "data", customEvent.Data["test"])
	assert.Equal(t, 123, customEvent.Data["id"])
}

// TestEventPlugin_SyncEvents 测试同步事件功能
func TestEventPlugin_SyncEvents(t *testing.T) {
	config := DefaultEventPluginConfig()
	plugin := NewEventPlugin("sync-test-plugin", config)

	ctx := context.Background()

	// 测试未启动状态
	_, err := plugin.SyncEvents(ctx, "test", time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// 启动插件
	err = plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 先获取并跳过startup事件
	_, err = plugin.SyncEvents(ctx, "test", time.Second)
	require.NoError(t, err)

	// 发布事件
	go func() {
		time.Sleep(time.Millisecond * 50)
		plugin.PublishEvent(EventTypeTick, map[string]interface{}{
			"tick_time": time.Now().Unix(),
		})
	}()

	// 同步获取事件
	event, err := plugin.SyncEvents(ctx, "test", time.Second)
	require.NoError(t, err)

	if e, ok := event.(Event); ok {
		assert.Equal(t, EventTypeTick, e.Type)
		assert.NotNil(t, e.Data["tick_time"])
	} else {
		t.Errorf("Expected Event type, got %T", event)
	}

	// 测试超时
	_, err = plugin.SyncEvents(ctx, "test", time.Millisecond*10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

// TestEventPlugin_SubscribeToEvents 测试事件订阅功能
func TestEventPlugin_SubscribeToEvents(t *testing.T) {
	config := DefaultEventPluginConfig()
	plugin := NewEventPlugin("subscribe-test-plugin", config)

	ctx := context.Background()

	// 测试未启动状态的订阅
	eventCh, unsubscribe := plugin.SubscribeToEvents(EventTypeTick)
	
	// 通道应该立即关闭
	select {
	case <-eventCh:
		// 这是期望的行为 - 通道已关闭
	case <-time.After(time.Millisecond * 100):
		t.Error("Expected channel to be closed immediately")
	}
	
	// 尝试发送事件到关闭的通道应该立即返回
	select {
	case <-eventCh:
		// 通道已关闭，这是正确的
	default:
	}
	unsubscribe()

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 订阅事件
	eventCh, unsubscribe = plugin.SubscribeToEvents(EventTypeTick)
	defer unsubscribe()

	// 发布不同类型的事件
	plugin.PublishEvent(EventTypeCustom, map[string]interface{}{"type": "custom"})
	plugin.PublishEvent(EventTypeTick, map[string]interface{}{"tick": 1})

	// 应该只收到订阅的事件类型
	select {
	case event := <-eventCh:
		assert.Equal(t, EventTypeTick, event.Type)
		assert.Equal(t, 1, event.Data["tick"])
	case <-time.After(time.Second):
		t.Error("Timeout waiting for subscribed event")
	}

	// 不应该收到自定义事件
	select {
	case <-eventCh:
		t.Error("Should not receive custom event")
	case <-time.After(time.Millisecond * 100):
		// 这是期望的行为
	}
}

// TestEventPlugin_Configuration 测试插件配置
func TestEventPlugin_Configuration(t *testing.T) {
	// 测试默认配置
	config := DefaultEventPluginConfig()
	assert.Equal(t, time.Second*5, config.TickInterval)
	assert.Equal(t, 100, config.EventBufferSize)
	assert.True(t, config.EnableLogging)

	// 测试自定义配置
	customConfig := EventPluginConfig{
		TickInterval:    time.Second * 10,
		EventBufferSize: 50,
		EnableLogging:   false,
	}

	plugin := NewEventPlugin("config-test-plugin", customConfig)
	retrievedConfig := plugin.GetConfig()

	assert.Equal(t, time.Second*10, retrievedConfig.TickInterval)
	assert.Equal(t, 50, retrievedConfig.EventBufferSize)
	assert.False(t, retrievedConfig.EnableLogging)
}

// TestEventPlugin_UpdateConfig 测试配置更新
func TestEventPlugin_UpdateConfig(t *testing.T) {
	config := DefaultEventPluginConfig()
	config.TickInterval = time.Millisecond * 500
	plugin := NewEventPlugin("update-config-plugin", config)

	ctx := context.Background()

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 更新配置
	newConfig := EventPluginConfig{
		TickInterval:    time.Second * 2,
		EventBufferSize: 200,
		EnableLogging:   false,
	}

	err = plugin.UpdateConfig(newConfig)
	require.NoError(t, err)

	updatedConfig := plugin.GetConfig()
	assert.Equal(t, time.Second*2, updatedConfig.TickInterval)
	assert.Equal(t, 200, updatedConfig.EventBufferSize)
	assert.False(t, updatedConfig.EnableLogging)
}

// TestEventPlugin_TickEvents 测试定时事件
func TestEventPlugin_TickEvents(t *testing.T) {
	config := DefaultEventPluginConfig()
	config.EventBufferSize = 2
	config.TickInterval = time.Millisecond * 50 // 快速测试
	plugin := NewEventPlugin("tick-test-plugin", config)

	ctx := context.Background()

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 监听事件
	eventCh := plugin.Events(ctx, nil)

	// 等待几个定时事件
	tickCount := 0
	timeout := time.After(time.Second * 2)

	for tickCount < 10 {
		select {
		case event := <-eventCh:
			if e, ok := event.(Event); ok && e.Type == EventTypeTick {
				tickCount++
			}
		case <-timeout:
			t.Errorf("Timeout waiting for tick events. Got %d ticks", tickCount)
			return
		}
	}

	assert.True(t, tickCount >= 3, "Should receive at least 3 tick events")
}

// TestEventPlugin_IntegrationWithPluginManager 测试与插件管理器的集成
func TestEventPlugin_IntegrationWithPluginManager(t *testing.T) {
	pm := NewPluginManager()
	ctx := context.Background()

	// 创建并注册事件插件
	config := DefaultEventPluginConfig()
	config.TickInterval = time.Millisecond * 100
	eventPlugin := NewEventPlugin("integration-test-plugin", config)

	err := pm.RegisterPlugin(eventPlugin)
	require.NoError(t, err)

	// 启动管理器（应该自动启动插件）
	err = pm.StartAll(ctx)
	require.NoError(t, err)
	defer pm.StopAll()

	// 验证插件正在运行
	assert.True(t, eventPlugin.IsRunning())

	// 通过插件管理器获取事件
	go func() {
		time.Sleep(time.Millisecond * 50)
		eventPlugin.PublishEvent(EventTypeCustom, map[string]interface{}{
			"integration_test": true,
		})
	}()

	// 需要获取两次事件：第一次是startup，第二次是custom
	// 跳过startup事件
	_, err = pm.GetSyncEvent("integration-test-plugin", "test", time.Second)
	require.NoError(t, err)

	// 获取custom事件
	event, err := pm.GetSyncEvent("integration-test-plugin", "test", time.Second)
	require.NoError(t, err)

	if e, ok := event.(Event); ok {
		assert.Equal(t, EventTypeCustom, e.Type)
		if val, exists := e.Data["integration_test"]; exists {
			assert.True(t, val.(bool))
		} else {
			t.Error("integration_test key not found in event data")
		}
	} else {
		t.Error("Expected Event type")
	}
}

// TestEventPlugin_ConcurrentOperations 测试并发操作
func TestEventPlugin_ConcurrentOperations(t *testing.T) {
	config := DefaultEventPluginConfig()
	config.EventBufferSize = 100
	plugin := NewEventPlugin("concurrent-test-plugin", config)

	ctx := context.Background()

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 并发发布事件
	done := make(chan bool, 2)

	// 启动发布协程
	go func() {
		for i := 0; i < 50; i++ {
			plugin.PublishEvent(EventTypeCustom, map[string]interface{}{
				"publisher": "goroutine1",
				"index":     i,
			})
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// 启动另一个发布协程
	go func() {
		for i := 0; i < 50; i++ {
			plugin.PublishEvent(EventTypeCustom, map[string]interface{}{
				"publisher": "goroutine2",
				"index":     i,
			})
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// 等待发布完成
	<-done
	<-done

	// 验证事件数量
	eventCh := plugin.Events(ctx, nil)
	eventCount := 0

	timeout := time.After(time.Second)
	for {
		select {
		case <-eventCh:
			eventCount++
		case <-timeout:
			assert.True(t, eventCount > 80, "Should receive most events")
			return
		}
	}
}

// BenchmarkEventPlugin_Publish 性能测试：事件发布
func BenchmarkEventPlugin_Publish(b *testing.B) {
	config := DefaultEventPluginConfig()
	config.EventBufferSize = 1000
	config.EnableLogging = false // 禁用日志以获得更准确的性能数据
	plugin := NewEventPlugin("benchmark-plugin", config)

	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin.PublishEvent(EventTypeCustom, map[string]interface{}{
			"index": i,
		})
	}
}
# Plugin Manager 事件插件库

支持事件发布、订阅、过滤和路由。

## 功能特性

- ✅ **插件管理器**：统一的插件注册、启动、停止和管理
- ✅ **基础事件插件**：简单易用的事件发布和订阅
- ✅ **高级事件插件**：支持事件过滤、优先级路由和统计
- ✅ **事件类型系统**：预定义事件类型和自定义事件支持
- ✅ **并发安全**：线程安全的事件处理
- ✅ **同步/异步事件**：支持同步和异步事件获取
- ✅ **生命周期管理**：完整的插件生命周期控制
- ✅ **统计功能**：事件统计和性能监控

## 快速开始

### 1. 基础事件插件

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/zmicro-team/ztlib/plugin_manager"
)

func main() {
    // 创建插件管理器
    pm := pluginmanager.NewPluginManager()
    
    // 创建基础事件插件
    config := pluginmanager.DefaultEventPluginConfig()
    config.TickInterval = time.Second * 2
    config.EnableLogging = true
    
    eventPlugin := pluginmanager.NewEventPlugin("my-plugin", config)
    
    // 注册并启动插件
    pm.RegisterPlugin(eventPlugin)
    ctx := context.Background()
    pm.StartAll(ctx)
    defer pm.StopAll()
    
    // 发布自定义事件
    eventPlugin.PublishEvent(pluginmanager.EventTypeCustom, map[string]interface{}{
        "message": "Hello World!",
        "user_id": 12345,
    })
    
    // 监听事件
    event, err := pm.GetSyncEvent("my-plugin", "request", time.Second*3)
    if err == nil {
        if e, ok := event.(pluginmanager.Event); ok {
            log.Printf("收到事件: %s - %v", e.Type, e.Data)
        }
    }
}
```

### 2. 事件订阅

```go
// 订阅特定类型的事件
eventCh, unsubscribe := eventPlugin.SubscribeToEvents(pluginmanager.EventTypeCustom)
defer unsubscribe()

// 监听订阅的事件
go func() {
    for event := range eventCh {
        log.Printf("订阅到事件: %s - %v", event.Type, event.Data)
    }
}()
```

### 3. 高级事件插件

```go
// 创建高级事件插件
config := pluginmanager.DefaultAdvancedEventPluginConfig()
config.EnableStats = true
advancedPlugin := pluginmanager.NewAdvancedEventPlugin("advanced-plugin", config)

// 添加过滤器（只允许自定义事件）
filter := pluginmanager.NewEventTypeFilter(pluginmanager.EventTypeCustom)
advancedPlugin.AddFilter(filter)

// 添加路由处理器
advancedPlugin.AddRoute("my-handler", func(event pluginmanager.Event) error {
    log.Printf("处理事件: %v", event.Data)
    return nil
}, 100) // 优先级

// 启动插件
ctx := context.Background()
advancedPlugin.Start(ctx)
defer advancedPlugin.Stop(ctx)

// 发布事件
err := advancedPlugin.PublishEvent(pluginmanager.Event{
    Type: pluginmanager.EventTypeCustom,
    Timestamp: time.Now(),
    Data: map[string]interface{}{"key": "value"},
    Source: "my-source",
})
```

## 事件类型

预定义的事件类型：

- `EventTypeStartup`：插件启动事件
- `EventTypeShutdown`：插件关闭事件  
- `EventTypeTick`：定时器事件
- `EventTypeCustom`：自定义事件

## 高级功能

### 事件过滤

```go
// 按事件类型过滤
typeFilter := pluginmanager.NewEventTypeFilter(
    pluginmanager.EventTypeCustom,
    pluginmanager.EventTypeTick,
)

// 按数据内容过滤
dataFilter := pluginmanager.NewDataFilter("user_id", "eq", 12345)

advancedPlugin.AddFilter(typeFilter)
advancedPlugin.AddFilter(dataFilter)
```

### 优先级路由

```go
// 高优先级路由（优先执行）
advancedPlugin.AddRoute("critical-handler", handleCritical, 1000)

// 中优先级路由
advancedPlugin.AddRoute("normal-handler", handleNormal, 500)

// 低优先级路由（最后执行）
advancedPlugin.AddRoute("logging-handler", handleLogging, 100)
```

### 事件统计

```go
// 获取统计信息
stats := advancedPlugin.GetStats()
log.Printf("总事件数: %d", stats.TotalEvents)
log.Printf("路由事件数: %d", stats.RoutedEvents)
log.Printf("事件类型: %v", stats.EventTypes)

// 重置统计
advancedPlugin.ResetStats()
```

## 配置选项

### 基础插件配置

```go
type EventPluginConfig struct {
    TickInterval    time.Duration // 定时事件间隔
    EventBufferSize int           // 事件通道缓冲大小
    EnableLogging  bool          // 是否启用日志
}
```

### 高级插件配置

```go
type AdvancedEventPluginConfig struct {
    EventBufferSize   int           // 事件缓冲大小
    EnableStats       bool          // 启用统计
    StatsInterval     time.Duration // 统计报告间隔
    DefaultPriority   int           // 默认路由优先级
    MaxRouteHandlers  int           // 最大路由处理器数量
    EnableLogging     bool          // 启用日志
}
```

## 插件管理器API

```go
// 创建管理器
pm := pluginmanager.NewPluginManager()

// 注册插件
pm.RegisterPlugin(plugin)

// 启动所有插件
pm.StartAll(ctx)

// 停止所有插件
pm.StopAll()

// 获取插件列表
plugins := pm.GetPlugins()

// 获取特定插件
plugin, exists := pm.GetPlugin("plugin-name")

// 同步获取事件
event, err := pm.GetSyncEvent("plugin-name", "request", timeout)

// 批量获取事件
events, err := pm.GetSyncEvents([]string{"plugin1", "plugin2"}, "request", timeout)

// 获取所有插件事件
allEvents, err := pm.GetAllSyncEvents("request", timeout)
```

## 运行示例

```bash
# 运行演示程序
go run ./plugin_manager/cmd/demo

# 运行测试
go test ./plugin_manager -v

# 运行特定测试
go test ./plugin_manager -run "TestEventPlugin" -v
```

## 文件结构

```
plugin_manager/
├── plugin_manager.go          # 核心插件管理器
├── plugin_manager_test.go     # 插件管理器测试
├── event_plugin.go           # 基础事件插件
├── event_plugin_test.go      # 基础事件插件测试
├── advanced_event_plugin.go   # 高级事件插件
├── advanced_example.go        # 高级插件使用示例
├── example_usage.go          # 基础插件使用示例
├── cmd/demo/main.go         # 完整演示程序
└── README.md               # 说明文档
```

## 性能特点

- 🚀 高并发支持：多个goroutine可安全并发使用
- 📊 内存高效：可配置的缓冲区大小
- ⚡ 低延迟：直接的事件通道传递
- 🔍 可观测性：内置统计和日志功能

## 最佳实践

1. **合理设置缓冲区大小**：根据事件频率调整EventBufferSize
2. **使用过滤器**：在高级插件中使用过滤器减少不必要的事件处理
3. **优先级设计**：合理设置路由处理器优先级
4. **错误处理**：在路由处理器中妥善处理错误
5. **资源清理**：使用defer确保插件正确停止

## 许可证

本项目遵循项目整体许可证。
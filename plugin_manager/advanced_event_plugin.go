package pluginmanager

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventFilter 事件过滤器接口
type EventFilter interface {
	Match(event Event) bool
	Description() string
}

// RouteTarget 路由目标
type RouteTarget struct {
	Name     string
	Priority int
	Handler  func(event Event) error
}

// AdvancedEventPlugin 高级事件插件，支持过滤和路由
type AdvancedEventPlugin struct {
	name     string
	running  bool
	events   chan Event
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	config   AdvancedEventPluginConfig
	filters  []EventFilter
	routes   []RouteTarget
	filterMu sync.RWMutex
	routeMu  sync.RWMutex
	stats    EventStats
	statsMu  sync.RWMutex
}

// EventStats 事件统计
type EventStats struct {
	TotalEvents    int64
	FilteredEvents int64
	RoutedEvents   int64
	FailedRoutes   int64
	LastEventTime  time.Time
	EventTypes     map[EventType]int64
}

// AdvancedEventPluginConfig 高级事件插件配置
type AdvancedEventPluginConfig struct {
	EventBufferSize  int           `json:"event_buffer_size"`  // 事件通道缓冲大小
	EnableStats      bool          `json:"enable_stats"`       // 是否启用统计
	StatsInterval    time.Duration `json:"stats_interval"`     // 统计报告间隔
	DefaultPriority  int           `json:"default_priority"`   // 默认路由优先级
	MaxRouteHandlers int           `json:"max_route_handlers"` // 最大路由处理器数量
	EnableLogging    bool          `json:"enable_logging"`     // 是否启用日志
}

// DefaultAdvancedEventPluginConfig 默认高级配置
func DefaultAdvancedEventPluginConfig() AdvancedEventPluginConfig {
	return AdvancedEventPluginConfig{
		EventBufferSize:  100,
		EnableStats:      true,
		StatsInterval:    time.Minute * 5,
		DefaultPriority:  100,
		MaxRouteHandlers: 50,
		EnableLogging:    false,
	}
}

// NewAdvancedEventPlugin 创建高级事件插件
func NewAdvancedEventPlugin(name string, config AdvancedEventPluginConfig) *AdvancedEventPlugin {
	if config.EventBufferSize <= 0 {
		config.EventBufferSize = 100
	}

	plugin := &AdvancedEventPlugin{
		name:    name,
		events:  make(chan Event, config.EventBufferSize),
		filters: make([]EventFilter, 0),
		routes:  make([]RouteTarget, 0),
		config:  config,
		stats: EventStats{
			EventTypes: make(map[EventType]int64),
		},
	}

	return plugin
}

// Start 启动高级事件插件
func (aep *AdvancedEventPlugin) Start(ctx context.Context) error {
	aep.mu.Lock()
	defer aep.mu.Unlock()

	if aep.running {
		return fmt.Errorf("advanced event plugin '%s' is already running", aep.name)
	}

	aep.ctx, aep.cancel = context.WithCancel(ctx)
	aep.running = true

	// 启动事件处理循环
	go aep.processEvents()

	// 启动统计报告
	if aep.config.EnableStats && aep.config.StatsInterval > 0 {
		go aep.statsReporter()
	}

	return nil
}

// Stop 停止高级事件插件
func (aep *AdvancedEventPlugin) Stop(ctx context.Context) error {
	aep.mu.Lock()
	defer aep.mu.Unlock()

	if !aep.running {
		return fmt.Errorf("advanced event plugin '%s' is not running", aep.name)
	}

	if aep.cancel != nil {
		aep.cancel()
	}

	close(aep.events)
	aep.running = false

	return nil
}

// Events 返回事件通道
func (aep *AdvancedEventPlugin) Events(ctx context.Context, value any) <-chan any {
	aep.mu.RLock()
	defer aep.mu.RUnlock()

	if !aep.running {
		ch := make(chan any)
		close(ch)
		return ch
	}

	resultCh := make(chan any, aep.config.EventBufferSize)
	go func() {
		defer close(resultCh)
		for {
			select {
			case event, ok := <-aep.events:
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
func (aep *AdvancedEventPlugin) SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error) {
	if !aep.IsRunning() {
		return nil, fmt.Errorf("plugin '%s' is not running", aep.name)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case event := <-aep.events:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Name 返回插件名称
func (aep *AdvancedEventPlugin) Name() string {
	return aep.name
}

// IsRunning 检查插件是否正在运行
func (aep *AdvancedEventPlugin) IsRunning() bool {
	aep.mu.RLock()
	defer aep.mu.RUnlock()
	return aep.running
}

// PublishEvent 发布事件
func (aep *AdvancedEventPlugin) PublishEvent(event Event) error {
	if !aep.IsRunning() {
		return fmt.Errorf("plugin '%s' is not running", aep.name)
	}

	select {
	case aep.events <- event:
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// AddFilter 添加事件过滤器
func (aep *AdvancedEventPlugin) AddFilter(filter EventFilter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}

	aep.filterMu.Lock()
	defer aep.filterMu.Unlock()

	aep.filters = append(aep.filters, filter)
	return nil
}

// RemoveFilter 移除事件过滤器
func (aep *AdvancedEventPlugin) RemoveFilter(filterDescription string) int {
	aep.filterMu.Lock()
	defer aep.filterMu.Unlock()

	removed := 0
	newFilters := make([]EventFilter, 0, len(aep.filters))

	for _, filter := range aep.filters {
		if filter.Description() != filterDescription {
			newFilters = append(newFilters, filter)
		} else {
			removed++
		}
	}

	aep.filters = newFilters
	return removed
}

// AddRoute 添加事件路由
func (aep *AdvancedEventPlugin) AddRoute(name string, handler func(Event) error, priority int) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	if priority <= 0 {
		priority = aep.config.DefaultPriority
	}

	aep.routeMu.Lock()
	defer aep.routeMu.Unlock()

	if len(aep.routes) >= aep.config.MaxRouteHandlers {
		return fmt.Errorf("maximum number of route handlers reached")
	}

	route := RouteTarget{
		Name:     name,
		Priority: priority,
		Handler:  handler,
	}

	// 按优先级插入
	inserted := false
	for i, existing := range aep.routes {
		if priority > existing.Priority {
			aep.routes = append(aep.routes[:i], append([]RouteTarget{route}, aep.routes[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		aep.routes = append(aep.routes, route)
	}

	return nil
}

// RemoveRoute 移除事件路由
func (aep *AdvancedEventPlugin) RemoveRoute(name string) bool {
	aep.routeMu.Lock()
	defer aep.routeMu.Unlock()

	for i, route := range aep.routes {
		if route.Name == name {
			aep.routes = append(aep.routes[:i], aep.routes[i+1:]...)
			return true
		}
	}

	return false
}

// GetStats 获取事件统计
func (aep *AdvancedEventPlugin) GetStats() EventStats {
	aep.statsMu.RLock()
	defer aep.statsMu.RUnlock()

	// 返回统计数据的副本
	stats := EventStats{
		TotalEvents:    aep.stats.TotalEvents,
		FilteredEvents: aep.stats.FilteredEvents,
		RoutedEvents:   aep.stats.RoutedEvents,
		FailedRoutes:   aep.stats.FailedRoutes,
		LastEventTime:  aep.stats.LastEventTime,
		EventTypes:     make(map[EventType]int64),
	}

	for k, v := range aep.stats.EventTypes {
		stats.EventTypes[k] = v
	}

	return stats
}

// ResetStats 重置统计数据
func (aep *AdvancedEventPlugin) ResetStats() {
	aep.statsMu.Lock()
	defer aep.statsMu.Unlock()

	aep.stats = EventStats{
		EventTypes: make(map[EventType]int64),
	}
}

// 内部方法：处理事件
func (aep *AdvancedEventPlugin) processEvents() {
	for {
		select {
		case event, ok := <-aep.events:
			if !ok {
				return
			}
			aep.handleEvent(event)
		case <-aep.ctx.Done():
			return
		}
	}
}

// 内部方法：处理单个事件
func (aep *AdvancedEventPlugin) handleEvent(event Event) {
	// 更新统计
	if aep.config.EnableStats {
		aep.updateStats(event)
	}

	// 应用过滤器
	aep.filterMu.RLock()
	filters := make([]EventFilter, len(aep.filters))
	copy(filters, aep.filters)
	aep.filterMu.RUnlock()

	for _, filter := range filters {
		if !filter.Match(event) {
			if aep.config.EnableStats {
				aep.statsMu.Lock()
				aep.stats.FilteredEvents++
				aep.statsMu.Unlock()
			}
			return
		}
	}

	// 路由事件
	aep.routeMu.RLock()
	routes := make([]RouteTarget, len(aep.routes))
	copy(routes, aep.routes)
	aep.routeMu.RUnlock()

	for _, route := range routes {
		err := route.Handler(event)
		if err != nil && aep.config.EnableLogging {
			// 记录路由失败
			if aep.config.EnableStats {
				aep.statsMu.Lock()
				aep.stats.FailedRoutes++
				aep.statsMu.Unlock()
			}
		} else if err == nil && aep.config.EnableStats {
			aep.statsMu.Lock()
			aep.stats.RoutedEvents++
			aep.statsMu.Unlock()
		}
	}
}

// 内部方法：更新统计
func (aep *AdvancedEventPlugin) updateStats(event Event) {
	aep.statsMu.Lock()
	defer aep.statsMu.Unlock()

	aep.stats.TotalEvents++
	aep.stats.LastEventTime = time.Now()
	aep.stats.EventTypes[event.Type]++
}

// 内部方法：统计报告器
func (aep *AdvancedEventPlugin) statsReporter() {
	ticker := time.NewTicker(aep.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := aep.GetStats()
			if aep.config.EnableLogging {
				fmt.Printf("Advanced Event Plugin '%s' Stats:\n", aep.name)
				fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
				fmt.Printf("  Filtered Events: %d\n", stats.FilteredEvents)
				fmt.Printf("  Routed Events: %d\n", stats.RoutedEvents)
				fmt.Printf("  Failed Routes: %d\n", stats.FailedRoutes)
				fmt.Printf("  Event Types: %v\n", stats.EventTypes)
			}
		case <-aep.ctx.Done():
			return
		}
	}
}

// EventTypeFilter 事件类型过滤器
type EventTypeFilter struct {
	AllowedTypes []EventType
}

func (etf *EventTypeFilter) Match(event Event) bool {
	if len(etf.AllowedTypes) == 0 {
		return true // 如果没有指定类型，则允许所有
	}

	for _, allowedType := range etf.AllowedTypes {
		if event.Type == allowedType {
			return true
		}
	}
	return false
}

func (etf *EventTypeFilter) Description() string {
	return fmt.Sprintf("EventTypeFilter: %v", etf.AllowedTypes)
}

// NewEventTypeFilter 创建事件类型过滤器
func NewEventTypeFilter(types ...EventType) *EventTypeFilter {
	return &EventTypeFilter{AllowedTypes: types}
}

// DataFilter 数据过滤器
type DataFilter struct {
	Key      string
	Value    interface{}
	Operator string // "eq", "ne", "gt", "lt", "contains"
}

func (df *DataFilter) Match(event Event) bool {
	value, exists := event.Data[df.Key]
	if !exists {
		return false
	}

	switch df.Operator {
	case "eq":
		return value == df.Value
	case "ne":
		return value != df.Value
	case "contains":
		if str, ok := value.(string); ok {
			if substr, ok := df.Value.(string); ok {
				return contains(str, substr)
			}
		}
		return false
	default:
		return false
	}
}

func (df *DataFilter) Description() string {
	return fmt.Sprintf("DataFilter: %s %s %v", df.Key, df.Operator, df.Value)
}

// NewDataFilter 创建数据过滤器
func NewDataFilter(key, operator string, value interface{}) *DataFilter {
	return &DataFilter{
		Key:      key,
		Value:    value,
		Operator: operator,
	}
}

// 辅助函数：检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

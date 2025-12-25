package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	pluginmanager "github.com/zmicro-team/ztlib/plugin_manager"
)

const (
	EventTypeConnectionAdd    pluginmanager.EventType = "connection_add"
	EventTypeConnectionRemove pluginmanager.EventType = "connection_remove"
	EventTypeConnectionUpdate pluginmanager.EventType = "connection_update"
)

// ConnectionManagerPlugin 连接管理器插件接口
type ConnectionManagerPlugin interface {
	pluginmanager.CollectorPlugin
	// 获取活跃连接数
	GetActiveCount() int
	// 根据用户ID获取连接信息（只读视图）
	GetConnectionByUserID(userID string) (UserConnectionView, bool)
	// 更新用户活跃时间
	UpdateActiveTime(userID string) bool
	// 清理不活跃连接
	CleanupInactive(timeout time.Duration) []string
	// 获取所有连接的只读视图
	GetAllConnectionsView() []UserConnectionView
	// 发布自定义连接事件
	PublishConnectionEvent(eventType pluginmanager.EventType, data map[string]interface{}) error
}

// UserConnectionView 只读的用户连接视图
type UserConnectionView struct {
	UserID     string    `json:"user_id"`     // 用户唯一标识
	LoginTime  time.Time `json:"login_time"`  // 登录时间
	LastActive time.Time `json:"last_active"` // 最后活跃时间
	IPAddress  string    `json:"ip_address"`  // 客户端IP地址
}

// UserConnection 表示一个已登录用户的长连接
type UserConnection struct {
	UserID     string                 // 用户唯一标识
	Connection interface{}            // 连接对象（如*websocket.Conn）
	LoginTime  time.Time              // 登录时间
	LastActive time.Time              // 最后活跃时间
	IPAddress  string                 // 客户端IP地址
	MetaData   map[string]interface{} // 扩展元数据
	metaMu     sync.RWMutex           // 元数据锁
}

// ConnectionManagerConfig 连接管理器配置
type ConnectionManagerConfig struct {
	EventBufferSize int           `json:"event_buffer_size"` // 事件通道缓冲大小
	EnableLogging   bool          `json:"enable_logging"`    // 是否启用日志
	AutoCleanup     bool          `json:"auto_cleanup"`      // 是否自动清理
	CleanupInterval time.Duration `json:"cleanup_interval"`  // 清理间隔
	InactiveTimeout time.Duration `json:"inactive_timeout"`  // 不活跃超时
}

// ConnectionManagerPluginImpl 连接管理器插件实现
type ConnectionManagerPluginImpl struct {
	name             string
	running          bool
	events           chan pluginmanager.Event
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	eventSubscribers map[string]chan pluginmanager.Event
	subMu            sync.RWMutex
	config           ConnectionManagerConfig

	//  添加连接管理器的核心功能字段
	connections  map[string]*UserConnection
	reverseIndex map[interface{}]string
	connMu       sync.RWMutex
}

// DefaultConnectionManagerConfig 默认配置
func DefaultConnectionManagerConfig() ConnectionManagerConfig {
	return ConnectionManagerConfig{
		EventBufferSize: 100,
		EnableLogging:   true,
		AutoCleanup:     true,
		CleanupInterval: time.Minute * 5,
		InactiveTimeout: time.Minute * 30,
	}
}

// NewConnectionManagerPlugin 创建新的连接管理器插件
func NewConnectionManagerPlugin(name string, config ConnectionManagerConfig) *ConnectionManagerPluginImpl {
	if config.EventBufferSize <= 0 {
		config.EventBufferSize = 100
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Minute * 5
	}
	if config.InactiveTimeout <= 0 {
		config.InactiveTimeout = time.Minute * 30
	}

	return &ConnectionManagerPluginImpl{
		name:             name,
		events:           make(chan pluginmanager.Event, config.EventBufferSize),
		eventSubscribers: make(map[string]chan pluginmanager.Event),
		config:           config,
		connections:      make(map[string]*UserConnection),
		reverseIndex:     make(map[interface{}]string),
	}
}

// Start 启动连接管理器插件
func (cmp *ConnectionManagerPluginImpl) Start(ctx context.Context) error {
	cmp.mu.Lock()
	defer cmp.mu.Unlock()

	if cmp.running {
		return fmt.Errorf("connection manager plugin '%s' is already running", cmp.name)
	}

	cmp.ctx, cmp.cancel = context.WithCancel(ctx)
	cmp.running = true

	// 启动事件处理器
	// go cmp.eventLoop()

	// 启动自动清理协程
	if cmp.config.AutoCleanup && cmp.config.CleanupInterval > 0 {
		go cmp.autoCleanupLoop()
	}

	// 发送启动事件
	cmp.sendEvent(pluginmanager.Event{
		Type:      pluginmanager.EventTypeStartup,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "Connection manager plugin started"},
		Source:    cmp.name,
	})

	if cmp.config.EnableLogging {
		log.Printf("Connection manager plugin '%s' started", cmp.name)
	}

	return nil
}

// Stop 停止连接管理器插件
func (cmp *ConnectionManagerPluginImpl) Stop(ctx context.Context) error {
	cmp.mu.Lock()
	defer cmp.mu.Unlock()

	if !cmp.running {
		return fmt.Errorf("connection manager plugin '%s' is not running", cmp.name)
	}

	// 发送关闭事件
	cmp.sendEvent(pluginmanager.Event{
		Type:      pluginmanager.EventTypeShutdown,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "Connection manager plugin stopping"},
		Source:    cmp.name,
	})

	// 取消上下文
	if cmp.cancel != nil {
		cmp.cancel()
	}

	// 关闭所有订阅者通道
	cmp.subMu.Lock()
	for id, ch := range cmp.eventSubscribers {
		close(ch)
		delete(cmp.eventSubscribers, id)
	}
	cmp.subMu.Unlock()

	// 关闭事件通道
	close(cmp.events)
	cmp.running = false

	if cmp.config.EnableLogging {
		log.Printf("Connection manager plugin '%s' stopped", cmp.name)
	}

	return nil
}

// Events 返回事件通道
func (cmp *ConnectionManagerPluginImpl) Events(ctx context.Context, value any) <-chan any {
	cmp.mu.RLock()
	defer cmp.mu.RUnlock()

	if !cmp.running {
		// 返回已关闭的通道
		ch := make(chan any)
		close(ch)
		return ch
	}

	// CHANGED: 为每个订阅者创建专用的事件转发通道
	resultCh := make(chan any, cmp.config.EventBufferSize)
	go func() {
		defer close(resultCh)
		for {
			select {
			case event, ok := <-cmp.events:
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
func (cmp *ConnectionManagerPluginImpl) SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error) {
	if !cmp.IsRunning() {
		return nil, fmt.Errorf("plugin '%s' is not running", cmp.name)
	}

	// 如果指定了超时，创建带超时的 context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case event := <-cmp.events:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Name 返回插件名称
func (cmp *ConnectionManagerPluginImpl) Name() string {
	return cmp.name
}

// IsRunning 检查插件是否正在运行
func (cmp *ConnectionManagerPluginImpl) IsRunning() bool {
	cmp.mu.RLock()
	defer cmp.mu.RUnlock()
	return cmp.running
}

// Add 添加或更新用户连接
func (cmp *ConnectionManagerPluginImpl) Add(userID string, conn interface{}, ip string) error {
	cmp.connMu.Lock()
	defer cmp.connMu.Unlock()

	now := time.Now()

	// 如果用户已有连接，先移除旧连接
	if oldConn, exists := cmp.connections[userID]; exists {
		delete(cmp.reverseIndex, oldConn.Connection)

		// CHANGED: 发布连接更新事件
		cmp.sendEvent(pluginmanager.Event{
			Type:      EventTypeConnectionUpdate,
			Timestamp: now,
			Data: map[string]interface{}{
				"user_id":    userID,
				"old_conn":   oldConn.Connection,
				"new_conn":   conn,
				"ip_address": ip,
				"action":     "replaced",
			},
			Source: cmp.name,
		})
	}

	// 创建新连接记录
	uc := &UserConnection{
		UserID:     userID,
		Connection: conn,
		LoginTime:  now,
		LastActive: now,
		IPAddress:  ip,
		MetaData:   make(map[string]interface{}),
	}

	cmp.connections[userID] = uc
	cmp.reverseIndex[conn] = userID

	// CHANGED: 发布连接添加事件
	cmp.sendEvent(pluginmanager.Event{
		Type:      EventTypeConnectionAdd,
		Timestamp: now,
		Data: map[string]interface{}{
			"user_id":    userID,
			"login_time": now,
			"ip_address": ip,
		},
		Source: cmp.name,
	})

	return nil
}

// RemoveByUserID 根据用户ID移除连接
func (cmp *ConnectionManagerPluginImpl) RemoveByUserID(userID string) bool {
	cmp.connMu.Lock()
	defer cmp.connMu.Unlock()

	if uc, exists := cmp.connections[userID]; exists {
		delete(cmp.reverseIndex, uc.Connection)
		delete(cmp.connections, userID)

		// CHANGED: 发布连接移除事件
		cmp.sendEvent(pluginmanager.Event{
			Type:      EventTypeConnectionRemove,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"user_id":      userID,
				"login_time":   uc.LoginTime,
				"last_active":  uc.LastActive,
				"ip_address":   uc.IPAddress,
				"duration_sec": time.Since(uc.LoginTime).Seconds(),
			},
			Source: cmp.name,
		})
		return true
	}
	return false
}

// RemoveByConnection 根据连接对象移除
func (cmp *ConnectionManagerPluginImpl) RemoveByConnection(conn interface{}) bool {
	cmp.connMu.Lock()
	defer cmp.connMu.Unlock()

	if userID, exists := cmp.reverseIndex[conn]; exists {
		if uc, ok := cmp.connections[userID]; ok {
			delete(cmp.connections, userID)
			delete(cmp.reverseIndex, conn)

			// CHANGED: 发布连接移除事件
			cmp.sendEvent(pluginmanager.Event{
				Type:      EventTypeConnectionRemove,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"user_id":      userID,
					"login_time":   uc.LoginTime,
					"last_active":  uc.LastActive,
					"ip_address":   uc.IPAddress,
					"duration_sec": time.Since(uc.LoginTime).Seconds(),
				},
				Source: cmp.name,
			})
		}
		return true
	}
	return false
}

// GetConnectionByUserID 根据用户ID获取连接信息（只读视图）
func (cmp *ConnectionManagerPluginImpl) GetConnectionByUserID(userID string) (UserConnectionView, bool) {
	cmp.connMu.RLock()
	defer cmp.connMu.RUnlock()

	uc, exists := cmp.connections[userID]
	if !exists {
		return UserConnectionView{}, false
	}

	return UserConnectionView{
		UserID:     uc.UserID,
		LoginTime:  uc.LoginTime,
		LastActive: uc.LastActive,
		IPAddress:  uc.IPAddress,
	}, true
}

// GetConnectionByConn 根据连接对象获取用户信息
func (cmp *ConnectionManagerPluginImpl) GetConnectionByConn(conn interface{}) (UserConnectionView, bool) {
	cmp.connMu.RLock()
	defer cmp.connMu.RUnlock()

	userID, exists := cmp.reverseIndex[conn]
	if !exists {
		return UserConnectionView{}, false
	}

	uc, exists := cmp.connections[userID]
	if !exists {
		return UserConnectionView{}, false
	}

	return UserConnectionView{
		UserID:     uc.UserID,
		LoginTime:  uc.LoginTime,
		LastActive: uc.LastActive,
		IPAddress:  uc.IPAddress,
	}, true
}

// UpdateActiveTime 更新用户最后活跃时间
func (cmp *ConnectionManagerPluginImpl) UpdateActiveTime(userID string) bool {
	cmp.connMu.Lock()
	defer cmp.connMu.Unlock()

	if uc, exists := cmp.connections[userID]; exists {
		oldActive := uc.LastActive
		uc.LastActive = time.Now()

		// CHANGED: 发布连接更新事件
		cmp.sendEvent(pluginmanager.Event{
			Type:      EventTypeConnectionUpdate,
			Timestamp: uc.LastActive,
			Data: map[string]interface{}{
				"user_id":          userID,
				"old_last_active":  oldActive,
				"new_last_active":  uc.LastActive,
				"inactive_seconds": uc.LastActive.Sub(oldActive).Seconds(),
			},
			Source: cmp.name,
		})
		return true
	}
	return false
}

// GetActiveCount 获取活跃连接数
func (cmp *ConnectionManagerPluginImpl) GetActiveCount() int {
	cmp.connMu.RLock()
	defer cmp.connMu.RUnlock()
	return len(cmp.connections)
}

// GetAllConnectionsView 获取所有连接的只读视图
func (cmp *ConnectionManagerPluginImpl) GetAllConnectionsView() []UserConnectionView {
	cmp.connMu.RLock()
	defer cmp.connMu.RUnlock()

	conns := make([]UserConnectionView, 0, len(cmp.connections))
	for _, uc := range cmp.connections {
		conns = append(conns, UserConnectionView{
			UserID:     uc.UserID,
			LoginTime:  uc.LoginTime,
			LastActive: uc.LastActive,
			IPAddress:  uc.IPAddress,
		})
	}
	return conns
}

// CleanupInactive 清理指定时间内不活跃的连接
func (cmp *ConnectionManagerPluginImpl) CleanupInactive(timeout time.Duration) []string {
	cmp.connMu.Lock()
	defer cmp.connMu.Unlock()

	var removed []string
	cutoff := time.Now().Add(-timeout)

	for userID, uc := range cmp.connections {
		if uc.LastActive.Before(cutoff) {
			delete(cmp.reverseIndex, uc.Connection)
			delete(cmp.connections, userID)
			removed = append(removed, userID)

			// CHANGED: 发布连接清理事件
			cmp.sendEvent(pluginmanager.Event{
				Type:      EventTypeConnectionRemove,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"user_id":      userID,
					"reason":       "inactive_timeout",
					"timeout_sec":  timeout.Seconds(),
					"last_active":  uc.LastActive,
					"inactive_sec": time.Since(uc.LastActive).Seconds(),
				},
				Source: cmp.name,
			})
		}
	}
	return removed
}

// PublishConnectionEvent 发布自定义连接事件
func (cmp *ConnectionManagerPluginImpl) PublishConnectionEvent(eventType pluginmanager.EventType, data map[string]interface{}) error {
	if !cmp.IsRunning() {
		return fmt.Errorf("plugin '%s' is not running", cmp.name)
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	event := pluginmanager.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Source:    cmp.name,
	}

	cmp.sendEvent(event)
	return nil
}

// SetUserMetadata 设置用户元数据（直接方法调用）
func (cmp *ConnectionManagerPluginImpl) SetUserMetadata(userID, key string, value interface{}) bool {
	cmp.connMu.RLock()
	uc, exists := cmp.connections[userID]
	cmp.connMu.RUnlock()

	if !exists {
		return false
	}

	uc.metaMu.Lock()
	if uc.MetaData == nil {
		uc.MetaData = make(map[string]interface{})
	}
	uc.MetaData[key] = value
	uc.metaMu.Unlock()

	return true
}

// GetUserMetadata 获取用户元数据（直接方法调用）
func (cmp *ConnectionManagerPluginImpl) GetUserMetadata(userID, key string) (interface{}, bool) {
	cmp.connMu.RLock()
	uc, exists := cmp.connections[userID]
	cmp.connMu.RUnlock()

	if !exists {
		return nil, false
	}

	uc.metaMu.RLock()
	value, exists := uc.MetaData[key]
	uc.metaMu.RUnlock()

	return value, exists
}

// 内部方法：发送事件
func (cmp *ConnectionManagerPluginImpl) sendEvent(event pluginmanager.Event) {
	select {
	case cmp.events <- event:
		// 事件发送成功，通知订阅者
		cmp.notifySubscribers(event)
		if cmp.config.EnableLogging {
			log.Printf("Connection event published: %+v", event)
		}
	default:
		// 事件通道已满，丢弃事件
		if cmp.config.EnableLogging {
			log.Printf("Connection event dropped: %+v (channel full)", event)
		}
	}
}

// 内部方法：通知订阅者
func (cmp *ConnectionManagerPluginImpl) notifySubscribers(event pluginmanager.Event) {
	cmp.subMu.RLock()
	defer cmp.subMu.RUnlock()

	for subscriberID, subscriberCh := range cmp.eventSubscribers {
		// 过滤事件类型：只发送连接相关事件给订阅者
		switch event.Type {
		case EventTypeConnectionAdd, EventTypeConnectionRemove, EventTypeConnectionUpdate:
			select {
			case subscriberCh <- event:
			default:
				// 订阅者通道已满，记录并跳过
				if cmp.config.EnableLogging {
					log.Printf("Subscriber %s channel full, event dropped", subscriberID)
				}
			}
		}
	}
}

// 内部方法：事件循环(拓展使用)
func (cmp *ConnectionManagerPluginImpl) eventLoop() {
	for {
		select {
		case <-cmp.ctx.Done():
			return
		}
	}
}

// 内部方法：自动清理循环
func (cmp *ConnectionManagerPluginImpl) autoCleanupLoop() {
	if cmp.config.CleanupInterval <= 0 || cmp.config.InactiveTimeout <= 0 {
		return
	}

	ticker := time.NewTicker(cmp.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			removed := cmp.CleanupInactive(cmp.config.InactiveTimeout)
			if len(removed) > 0 && cmp.config.EnableLogging {
				log.Printf("Auto cleanup removed %d inactive connections: %v", len(removed), removed)
			}
		case <-cmp.ctx.Done():
			return
		}
	}
}

// SubscribeToConnectionEvents 订阅连接事件（专用订阅方法）
func (cmp *ConnectionManagerPluginImpl) SubscribeToConnectionEvents() (<-chan pluginmanager.Event, func()) {
	if !cmp.IsRunning() {
		ch := make(chan pluginmanager.Event)
		close(ch)
		return ch, func() {}
	}

	cmp.subMu.Lock()
	defer cmp.subMu.Unlock()

	subscriberID := uuid.New().String()
	ch := make(chan pluginmanager.Event, 10) // 小缓冲区用于订阅

	// 将订阅信息存储
	cmp.eventSubscribers[subscriberID] = ch

	unsubscribe := func() {
		cmp.subMu.Lock()
		defer cmp.subMu.Unlock()
		if subCh, exists := cmp.eventSubscribers[subscriberID]; exists {
			close(subCh)
			delete(cmp.eventSubscribers, subscriberID)
		}
	}

	return ch, unsubscribe
}

// GetConfig 获取插件配置
func (cmp *ConnectionManagerPluginImpl) GetConfig() ConnectionManagerConfig {
	cmp.mu.RLock()
	defer cmp.mu.RUnlock()
	return cmp.config
}

// UpdateConfig 更新插件配置
func (cmp *ConnectionManagerPluginImpl) UpdateConfig(config ConnectionManagerConfig) error {
	cmp.mu.Lock()
	defer cmp.mu.Unlock()

	// 如果正在运行且清理间隔改变，需要重启清理协程
	if config.CleanupInterval != cmp.config.CleanupInterval && config.CleanupInterval > 0 {
		// 注意：这里简化处理，实际应该重新启动清理协程
		cmp.config.CleanupInterval = config.CleanupInterval
	}

	cmp.config = config
	return nil
}

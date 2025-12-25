package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockConnection 模拟连接对象
type MockConnection struct {
	ID   string
	Data string
}

// TestConnectionManagerPlugin_Basic 测试基本功能
func TestConnectionManagerPlugin_Basic(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("basic-test", config)
	ctx := context.Background()

	// 测试初始状态
	assert.Equal(t, "basic-test", plugin.Name())
	assert.False(t, plugin.IsRunning())
	assert.Equal(t, 0, plugin.GetActiveCount())

	// 启动插件
	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)
	assert.True(t, plugin.IsRunning())

	// 测试添加连接
	userID := "test-user"
	conn := &MockConnection{ID: uuid.New().String(), Data: "test"}
	ip := "192.168.1.100"

	err = plugin.Add(userID, conn, ip)
	require.NoError(t, err)
	assert.Equal(t, 1, plugin.GetActiveCount())

	// 验证连接信息
	connView, exists := plugin.GetConnectionByUserID(userID)
	require.True(t, exists)
	assert.Equal(t, userID, connView.UserID)
	assert.Equal(t, ip, connView.IPAddress)
	assert.False(t, connView.LoginTime.IsZero())

	// 测试更新活跃时间
	oldActive := connView.LastActive
	time.Sleep(time.Millisecond * 10)
	
	updated := plugin.UpdateActiveTime(userID)
	assert.True(t, updated)

	newConnView, _ := plugin.GetConnectionByUserID(userID)
	assert.True(t, newConnView.LastActive.After(oldActive))

	// 测试移除连接
	removed := plugin.RemoveByUserID(userID)
	assert.True(t, removed)
	assert.Equal(t, 0, plugin.GetActiveCount())

	// 验证连接已移除
	_, exists = plugin.GetConnectionByUserID(userID)
	assert.False(t, exists)
}

// TestConnectionManagerPlugin_Configuration 测试配置
func TestConnectionManagerPlugin_Configuration(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	assert.Equal(t, 100, config.EventBufferSize)
	assert.True(t, config.EnableLogging)
	assert.True(t, config.AutoCleanup)
	assert.Equal(t, time.Minute*5, config.CleanupInterval)
	assert.Equal(t, time.Minute*30, config.InactiveTimeout)
}

// TestConnectionManagerPlugin_CreateWithCustomConfig 测试自定义配置
func TestConnectionManagerPlugin_CreateWithCustomConfig(t *testing.T) {
	config := ConnectionManagerConfig{
		EventBufferSize: 200,
		EnableLogging:   false,
		AutoCleanup:     false,
		CleanupInterval: time.Minute * 10,
		InactiveTimeout:  time.Hour,
	}

	plugin := NewConnectionManagerPlugin("custom-config-test", config)
	
	finalConfig := plugin.GetConfig()
	assert.Equal(t, 200, finalConfig.EventBufferSize)
	assert.False(t, finalConfig.EnableLogging)
	assert.False(t, finalConfig.AutoCleanup)
	assert.Equal(t, time.Minute*10, finalConfig.CleanupInterval)
	assert.Equal(t, time.Hour, finalConfig.InactiveTimeout)
}

// TestConnectionManagerPlugin_ZeroConfig 测试零值配置处理
func TestConnectionManagerPlugin_ZeroConfig(t *testing.T) {
	config := ConnectionManagerConfig{
		EventBufferSize: 0, // 应该使用默认值
		CleanupInterval: 0, // 应该使用默认值
		InactiveTimeout: 0, // 应该使用默认值
	}

	plugin := NewConnectionManagerPlugin("zero-config-test", config)
	
	finalConfig := plugin.GetConfig()
	// 零值应该被替换为默认值
	assert.Equal(t, 100, finalConfig.EventBufferSize) // 默认缓冲区大小
	assert.Equal(t, time.Minute*5, finalConfig.CleanupInterval) // 默认清理间隔
	assert.Equal(t, time.Minute*30, finalConfig.InactiveTimeout) // 默认超时
}

// TestConnectionManagerPlugin_MultipleConnections 测试多个连接管理
func TestConnectionManagerPlugin_MultipleConnections(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("multi-conn-test", config)
	ctx := context.Background()

	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 添加多个连接
	users := []string{"user1", "user2", "user3"}
	connections := make([]*MockConnection, len(users))

	for i, userID := range users {
		connections[i] = &MockConnection{
			ID:   uuid.New().String(),
			Data: "data-" + userID,
		}
		ip := "192.168.1." + string(rune(100+i))
		
		err := plugin.Add(userID, connections[i], ip)
		require.NoError(t, err)
	}

	// 验证连接数
	assert.Equal(t, len(users), plugin.GetActiveCount())

	// 获取所有连接视图
	allConns := plugin.GetAllConnectionsView()
	assert.Len(t, allConns, len(users))

	// 验证每个用户都在列表中
	userMap := make(map[string]bool)
	for _, connView := range allConns {
		userMap[connView.UserID] = true
		assert.False(t, connView.LoginTime.IsZero())
		assert.NotEmpty(t, connView.IPAddress)
	}

	for _, userID := range users {
		assert.True(t, userMap[userID], "User %s should be in connections", userID)
	}

	// 根据连接对象移除连接
	removed := plugin.RemoveByConnection(connections[0])
	assert.True(t, removed)
	assert.Equal(t, len(users)-1, plugin.GetActiveCount())

	// 验证对应用户被移除
	_, exists := plugin.GetConnectionByUserID(users[0])
	assert.False(t, exists)
}

// TestConnectionManagerPlugin_CleanupInactive 测试清理不活跃连接
func TestConnectionManagerPlugin_CleanupInactive(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	config.AutoCleanup = false // 手动测试清理
	plugin := NewConnectionManagerPlugin("cleanup-test", config)
	ctx := context.Background()

	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 添加测试连接
	activeUser := "active-user"
	inactiveUser := "inactive-user"

	activeConn := &MockConnection{ID: "active", Data: "active"}
	inactiveConn := &MockConnection{ID: "inactive", Data: "inactive"}

	// 添加活跃用户
	err = plugin.Add(activeUser, activeConn, "192.168.1.100")
	require.NoError(t, err)

	// 添加不活跃用户
	err = plugin.Add(inactiveUser, inactiveConn, "192.168.1.101")
	require.NoError(t, err)

	// 更新活跃用户的活跃时间
	time.Sleep(time.Millisecond * 50)
	plugin.UpdateActiveTime(activeUser)

	// 清理不活跃的连接（设置短超时）
	timeout := time.Millisecond * 20
	cleanupResult := plugin.CleanupInactive(timeout)

	// 验证只有不活跃用户被清理
	assert.Len(t, cleanupResult, 1)
	assert.Contains(t, cleanupResult, inactiveUser)
	assert.Equal(t, 1, plugin.GetActiveCount())

	// 验证活跃用户仍然存在
	_, exists := plugin.GetConnectionByUserID(activeUser)
	assert.True(t, exists)

	// 验证不活跃用户已被移除
	_, exists = plugin.GetConnectionByUserID(inactiveUser)
	assert.False(t, exists)
}

// TestConnectionManagerPlugin_Metadata 测试元数据功能
func TestConnectionManagerPlugin_Metadata(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("metadata-test", config)
	ctx := context.Background()

	err := plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 添加连接
	userID := "metadata-user"
	conn := &MockConnection{ID: uuid.New().String(), Data: "metadata"}
	ip := "192.168.1.200"

	err = plugin.Add(userID, conn, ip)
	require.NoError(t, err)

	// 设置元数据
	success := plugin.SetUserMetadata(userID, "session_id", "session123")
	assert.True(t, success)
	
	success = plugin.SetUserMetadata(userID, "user_role", "admin")
	assert.True(t, success)

	// 获取元数据
	sessionID, exists := plugin.GetUserMetadata(userID, "session_id")
	require.True(t, exists)
	assert.Equal(t, "session123", sessionID)

	userRole, exists := plugin.GetUserMetadata(userID, "user_role")
	require.True(t, exists)
	assert.Equal(t, "admin", userRole)

	// 获取不存在的元数据
	_, exists = plugin.GetUserMetadata(userID, "nonexistent")
	assert.False(t, exists)

	// 为不存在的用户设置元数据
	success = plugin.SetUserMetadata("nonexistent", "key", "value")
	assert.False(t, success)
}

// TestConnectionManagerPlugin_Lifecycle 错误测试生命周期
func TestConnectionManagerPlugin_Lifecycle(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("lifecycle-test", config)
	ctx := context.Background()

	// 测试重复启动
	err := plugin.Start(ctx)
	require.NoError(t, err)
	
	err = plugin.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// 测试重复停止
	err = plugin.Stop(ctx)
	require.NoError(t, err)
	
	err = plugin.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// TestConnectionManagerPlugin_SyncEvents 测试同步事件
func TestConnectionManagerPlugin_SyncEvents(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("sync-test", config)
	ctx := context.Background()

	// 测试未启动状态
	_, err := plugin.SyncEvents(ctx, "test", time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// 启动插件
	err = plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 先获取startup事件
	_, err = plugin.SyncEvents(ctx, "test", time.Second)
	require.NoError(t, err)

	// 添加连接以产生事件
	go func() {
		time.Sleep(time.Millisecond * 50)
		userID := "sync-user"
		conn := &MockConnection{ID: "sync", Data: "sync-test"}
		plugin.Add(userID, conn, "192.168.1.999")
	}()

	// 获取连接事件
	event, err := plugin.SyncEvents(ctx, "test", time.Second*2)
	require.NoError(t, err)
	require.NotNil(t, event)
}

// TestConnectionManagerPlugin_PublishEvent 测试发布自定义事件
func TestConnectionManagerPlugin_PublishEvent(t *testing.T) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("publish-test", config)
	ctx := context.Background()

	// 测试未启动状态
	err := plugin.PublishConnectionEvent(EventTypeConnectionAdd, map[string]interface{}{
		"test": "data",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// 启动插件
	err = plugin.Start(ctx)
	require.NoError(t, err)
	defer plugin.Stop(ctx)

	// 发布自定义事件
	customData := map[string]interface{}{
		"custom_type": "test_event",
		"payload":    "test payload",
	}

	err = plugin.PublishConnectionEvent(EventTypeConnectionAdd, customData)
	require.NoError(t, err)
}

// BenchmarkConnectionManagerPlugin_Add 性能测试：添加连接
func BenchmarkConnectionManagerPlugin_Add(b *testing.B) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("bench-add", config)
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := "bench_user"
		conn := &MockConnection{
			ID:   "bench-" + string(rune(i%1000)),
			Data: "bench-data",
		}
		ip := "192.168.100.1"
		plugin.Add(userID, conn, ip)
	}
}

// BenchmarkConnectionManagerPlugin_Get 性能测试：获取连接
func BenchmarkConnectionManagerPlugin_Get(b *testing.B) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("bench-get", config)
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	// 预先添加连接
	userID := "bench_get_user"
	conn := &MockConnection{
		ID:   "bench-get-conn",
		Data: "bench-get-data",
	}
	ip := "192.168.200.1"
	plugin.Add(userID, conn, ip)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin.GetConnectionByUserID(userID)
	}
}

// BenchmarkConnectionManagerPlugin_UpdateActiveTime 性能测试：更新活跃时间
func BenchmarkConnectionManagerPlugin_UpdateActiveTime(b *testing.B) {
	config := DefaultConnectionManagerConfig()
	config.EnableLogging = false
	plugin := NewConnectionManagerPlugin("bench-update", config)
	ctx := context.Background()
	plugin.Start(ctx)
	defer plugin.Stop(ctx)

	// 预先添加连接
	userID := "bench_update_user"
	conn := &MockConnection{
		ID:   "bench-update-conn",
		Data: "bench-update-data",
	}
	ip := "192.168.300.1"
	plugin.Add(userID, conn, ip)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin.UpdateActiveTime(userID)
	}
}
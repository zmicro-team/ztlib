package pluginmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPlugin 实现 CollectorPlugin 接口用于测试
type MockPlugin struct {
	name     string
	running  bool
	startErr error
	stopErr  error
	events   chan any
	mu       sync.RWMutex
	ctx      context.Context
}

func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name:   name,
		events: make(chan any, 10),
	}
}

func (m *MockPlugin) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startErr != nil {
		return m.startErr
	}

	m.running = true
	m.ctx = ctx
	return nil
}

func (m *MockPlugin) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopErr != nil {
		return m.stopErr
	}

	m.running = false
	close(m.events)
	m.events = make(chan any, 10)
	return nil
}

func (m *MockPlugin) Events(ctx context.Context, value any) <-chan any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events
}

func (m *MockPlugin) Name() string {
	return m.name
}

func (m *MockPlugin) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

func (m *MockPlugin) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startErr = err
}

func (m *MockPlugin) SetStopError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopErr = err
}

func (m *MockPlugin) SendEvent(event any) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case m.events <- event:
	default:
	}
}

func (m *MockPlugin) SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error) {
	m.mu.RLock()
	running := m.running
	events := m.events
	m.mu.RUnlock()

	if !running {
		return nil, fmt.Errorf("plugin '%s' is not running", m.name)
	}

	// 如果指定了超时，创建带超时的 context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case event := <-events:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TestNewPluginManager 测试 NewPluginManager 工厂函数
func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()

	assert.NotNil(t, pm)
	assert.Empty(t, pm.GetPlugins())
	assert.False(t, pm.running)
	assert.Nil(t, pm.ctx)
	assert.Nil(t, pm.cancel)
}

// TestPluginManager_RegisterPlugin 测试插件注册功能
func TestPluginManager_RegisterPlugin(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*PluginManager)
		plugin      CollectorPlugin
		expectError bool
		errorMsg    string
	}{
		{
			name:        "成功注册新插件",
			plugin:      NewMockPlugin("test-plugin"),
			expectError: false,
		},
		{
			name: "注册重复插件名应该失败",
			setupFunc: func(pm *PluginManager) {
				plugin := NewMockPlugin("duplicate-plugin")
				err := pm.RegisterPlugin(plugin)
				require.NoError(t, err)
			},
			plugin:      NewMockPlugin("duplicate-plugin"),
			expectError: true,
			errorMsg:    "plugin with name 'duplicate-plugin' already registered",
		},
		{
			name: "管理器运行时自动启动新插件",
			setupFunc: func(pm *PluginManager) {
				ctx := context.Background()
				err := pm.StartAll(ctx)
				require.NoError(t, err)
			},
			plugin:      NewMockPlugin("auto-start-plugin"),
			expectError: false,
		},
		{
			name: "插件启动失败时应该回滚注册",
			setupFunc: func(pm *PluginManager) {
				ctx := context.Background()
				err := pm.StartAll(ctx)
				require.NoError(t, err)
			},
			plugin: func() CollectorPlugin {
				p := NewMockPlugin("fail-start-plugin")
				p.SetStartError(errors.New("start failed"))
				return p
			}(),
			expectError: true,
			errorMsg:    "failed to start plugin 'fail-start-plugin': start failed",
		},
		{
			name:        "成功注册事件插件",
			plugin:      NewEventPlugin("test-event-plugin", DefaultEventPluginConfig()),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPluginManager()

			if tt.setupFunc != nil {
				tt.setupFunc(pm)
			}

			err := pm.RegisterPlugin(tt.plugin)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// 验证插件已注册
				_, exists := pm.GetPlugin(tt.plugin.Name())
				assert.True(t, exists, "插件应该已注册")
			}
		})
	}
}

// TestPluginManager_UnregisterPlugin 测试插件注销功能
func TestPluginManager_UnregisterPlugin(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*PluginManager) string
		pluginName  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功注销已注册插件",
			setupFunc: func(pm *PluginManager) string {
				plugin := NewMockPlugin("unregister-test")
				err := pm.RegisterPlugin(plugin)
				require.NoError(t, err)
				return plugin.Name()
			},
			pluginName:  "unregister-test",
			expectError: false,
		},
		{
			name:        "注销不存在的插件应该失败",
			pluginName:  "non-existent",
			expectError: true,
			errorMsg:    "plugin with name 'non-existent' not found",
		},
		{
			name: "注销运行中的插件应该先停止",
			setupFunc: func(pm *PluginManager) string {
				plugin := NewMockPlugin("running-plugin")
				err := pm.RegisterPlugin(plugin)
				require.NoError(t, err)

				ctx := context.Background()
				err = pm.StartAll(ctx)
				require.NoError(t, err)
				return plugin.Name()
			},
			pluginName:  "running-plugin",
			expectError: false,
		},
		{
			name: "插件停止失败时应该返回错误",
			setupFunc: func(pm *PluginManager) string {
				plugin := NewMockPlugin("stop-fail-plugin")
				plugin.SetStopError(errors.New("stop failed"))
				err := pm.RegisterPlugin(plugin)
				require.NoError(t, err)

				ctx := context.Background()
				err = pm.StartAll(ctx)
				require.NoError(t, err)
				return plugin.Name()
			},
			pluginName:  "stop-fail-plugin",
			expectError: true,
			errorMsg:    "failed to stop plugin 'stop-fail-plugin': stop failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPluginManager()
			var actualPluginName string

			if tt.setupFunc != nil {
				actualPluginName = tt.setupFunc(pm)
			}

			if tt.pluginName == "" {
				tt.pluginName = actualPluginName
			}

			err := pm.UnregisterPlugin(tt.pluginName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// 验证插件已注销
				_, exists := pm.GetPlugin(tt.pluginName)
				assert.False(t, exists, "插件应该已注销")
			}
		})
	}
}

// TestPluginManager_StartAll 测试启动所有插件功能
func TestPluginManager_StartAll(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(*PluginManager)
		expectError     bool
		errorMsg        string
		expectedRunning bool
	}{
		{
			name: "成功启动所有插件",
			setupFunc: func(pm *PluginManager) {
				plugin1 := NewMockPlugin("plugin1")
				plugin2 := NewMockPlugin("plugin2")
				pm.RegisterPlugin(plugin1)
				pm.RegisterPlugin(plugin2)
			},
			expectError:     false,
			expectedRunning: true,
		},
		{
			name: "重复启动应该失败",
			setupFunc: func(pm *PluginManager) {
				plugin := NewMockPlugin("duplicate-start")
				pm.RegisterPlugin(plugin)

				ctx := context.Background()
				err := pm.StartAll(ctx)
				require.NoError(t, err)
			},
			expectError:     true,
			errorMsg:        "plugin manager is already running",
			expectedRunning: true, // 管理器应该仍然在运行状态
		},
		{
			name: "某个插件启动失败时应该停止所有插件",
			setupFunc: func(pm *PluginManager) {
				plugin1 := NewMockPlugin("good-plugin")
				plugin2 := NewMockPlugin("bad-plugin")
				plugin2.SetStartError(errors.New("start error"))
				pm.RegisterPlugin(plugin1)
				pm.RegisterPlugin(plugin2)
			},
			expectError: true,
			errorMsg:    "failed to start plugin 'bad-plugin': start error",
		},
		{
			name:            "空插件列表启动应该成功",
			expectError:     false,
			expectedRunning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPluginManager()

			if tt.setupFunc != nil {
				tt.setupFunc(pm)
			}

			ctx := context.Background()
			err := pm.StartAll(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				// 对于重复启动的情况，管理器应该仍然在运行状态
				if tt.expectedRunning {
					assert.True(t, pm.running, "管理器应该仍然在运行状态")
				} else {
					// 验证管理器未运行
					assert.False(t, pm.running)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, pm.running, "管理器应该正在运行")
				assert.NotNil(t, pm.ctx)
				assert.NotNil(t, pm.cancel)

				// 验证所有插件都在运行
				if tt.expectedRunning {
					for _, name := range pm.GetPlugins() {
						plugin, exists := pm.GetPlugin(name)
						require.True(t, exists)
						assert.True(t, plugin.IsRunning(), "插件 %s 应该正在运行", name)
					}
				}
			}
		})
	}
}

// TestPluginManager_StopAll 测试停止所有插件功能
func TestPluginManager_StopAll(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*PluginManager)
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功停止所有插件",
			setupFunc: func(pm *PluginManager) {
				plugin1 := NewMockPlugin("stop-plugin1")
				plugin2 := NewMockPlugin("stop-plugin2")
				pm.RegisterPlugin(plugin1)
				pm.RegisterPlugin(plugin2)

				ctx := context.Background()
				err := pm.StartAll(ctx)
				require.NoError(t, err)
			},
			expectError: false,
		},
		{
			name:        "停止未运行的管理器应该失败",
			expectError: true,
			errorMsg:    "plugin manager is not running",
		},
		{
			name: "部分插件停止失败应该继续停止其他插件并返回最后一个错误",
			setupFunc: func(pm *PluginManager) {
				plugin1 := NewMockPlugin("good-stop-plugin")
				plugin2 := NewMockPlugin("bad-stop-plugin")
				plugin3 := NewMockPlugin("another-bad-stop-plugin")
				plugin2.SetStopError(errors.New("stop error 1"))
				plugin3.SetStopError(errors.New("stop error 2"))
				pm.RegisterPlugin(plugin1)
				pm.RegisterPlugin(plugin2)
				pm.RegisterPlugin(plugin3)

				ctx := context.Background()
				err := pm.StartAll(ctx)
				require.NoError(t, err)
			},
			expectError: true,
			errorMsg:    "failed to stop plugin", // 更灵活的错误消息检查
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPluginManager()

			if tt.setupFunc != nil {
				tt.setupFunc(pm)
			}

			err := pm.StopAll()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证管理器状态
			assert.False(t, pm.running)
			assert.Nil(t, pm.ctx)
			assert.Nil(t, pm.cancel)
		})
	}
}

// TestPluginManager_GetPlugins 测试获取插件列表功能
func TestPluginManager_GetPlugins(t *testing.T) {
	pm := NewPluginManager()

	// 空列表
	assert.Empty(t, pm.GetPlugins())

	// 添加插件
	plugin1 := NewMockPlugin("get-plugin1")
	plugin2 := NewMockPlugin("get-plugin2")
	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)

	plugins := pm.GetPlugins()
	assert.Len(t, plugins, 2)
	assert.Contains(t, plugins, "get-plugin1")
	assert.Contains(t, plugins, "get-plugin2")
}

// TestPluginManager_GetPlugin 测试获取特定插件功能
func TestPluginManager_GetPlugin(t *testing.T) {
	pm := NewPluginManager()

	// 不存在的插件
	plugin, exists := pm.GetPlugin("non-existent")
	assert.Nil(t, plugin)
	assert.False(t, exists)

	// 存在的插件
	testPlugin := NewMockPlugin("get-test-plugin")
	pm.RegisterPlugin(testPlugin)

	retrievedPlugin, exists := pm.GetPlugin("get-test-plugin")
	assert.True(t, exists)
	assert.Equal(t, testPlugin, retrievedPlugin)
}

// TestPluginManager_ConcurrentAccess 测试并发访问安全性
func TestPluginManager_ConcurrentAccess(t *testing.T) {
	pm := NewPluginManager()
	ctx := context.Background()

	// 注册多个插件
	for i := 0; i < 10; i++ {
		plugin := NewMockPlugin(fmt.Sprintf("concurrent-plugin-%d", i))
		err := pm.RegisterPlugin(plugin)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// 并发启动
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := pm.StartAll(ctx); err != nil {
			errChan <- err
		}
	}()

	// 并发注册新插件
	wg.Add(5)
	for i := 10; i < 15; i++ {
		go func(id int) {
			defer wg.Done()
			plugin := NewMockPlugin(fmt.Sprintf("late-plugin-%d", id))
			if err := pm.RegisterPlugin(plugin); err != nil {
				errChan <- err
			}
		}(i)
	}

	// 并发读取
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			plugins := pm.GetPlugins()
			if len(plugins) == 0 {
				errChan <- errors.New("插件列表为空")
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("并发操作出现错误: %v", err)
	}

	// 清理
	pm.StopAll()
}

// TestPluginManager_Events 测试事件通道功能
func TestPluginManager_Events(t *testing.T) {
	pm := NewPluginManager()
	ctx := context.Background()

	// 创建并发送事件的插件
	plugin := NewMockPlugin("events-test-plugin")
	pm.RegisterPlugin(plugin)

	err := pm.StartAll(ctx)
	require.NoError(t, err)

	// 发送测试事件
	testEvent := "test-event-data"
	plugin.SendEvent(testEvent)

	// 检查事件是否正确发送
	select {
	case event := <-plugin.Events(ctx, nil):
		assert.Equal(t, testEvent, event)
	case <-time.After(100 * time.Millisecond):
		t.Error("未接收到预期的事件")
	}

	pm.StopAll()
}

// TestPluginManager_ZeroValue 测试零值情况
func TestPluginManager_ZeroValue(t *testing.T) {
	// 创建一个空的管理器来测试基本行为
	pm := NewPluginManager()

	// 空管理器应该可以安全调用方法
	assert.Empty(t, pm.GetPlugins())
	_, exists := pm.GetPlugin("any")
	assert.False(t, exists)

	// 停止未运行的管理器应该返回错误
	err := pm.StopAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// TestPluginManager_MultiplePluginsWithSameName 测试处理同名插件的边界情况
func TestPluginManager_MultiplePluginsWithSameName(t *testing.T) {
	pm := NewPluginManager()

	// 注册第一个插件
	plugin1 := NewMockPlugin("same-name")
	err := pm.RegisterPlugin(plugin1)
	assert.NoError(t, err)

	// 尝试注册同名插件应该失败
	plugin2 := NewMockPlugin("same-name")
	err = pm.RegisterPlugin(plugin2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// 验证第一个插件仍然存在
	retrieved, exists := pm.GetPlugin("same-name")
	assert.True(t, exists)
	assert.Equal(t, plugin1, retrieved)
}

// TestPluginManager_LifecycleConsistency 测试生命周期一致性
func TestPluginManager_LifecycleConsistency(t *testing.T) {
	pm := NewPluginManager()
	ctx := context.Background()

	// 注册插件
	plugin1 := NewMockPlugin("lifecycle-plugin1")
	plugin2 := NewMockPlugin("lifecycle-plugin2")
	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)

	// 启动所有插件
	err := pm.StartAll(ctx)
	require.NoError(t, err)

	// 验证所有插件都在运行
	assert.True(t, plugin1.IsRunning())
	assert.True(t, plugin2.IsRunning())
	assert.True(t, pm.running)

	// 停止所有插件
	err = pm.StopAll()
	assert.NoError(t, err)

	// 验证所有插件都已停止
	assert.False(t, plugin1.IsRunning())
	assert.False(t, plugin2.IsRunning())
	assert.False(t, pm.running)
	assert.Nil(t, pm.ctx)
	assert.Nil(t, pm.cancel)
}

// TestPluginManager_ErrorHandling 测试错误处理的健壮性
func TestPluginManager_ErrorHandling(t *testing.T) {
	pm := NewPluginManager()
	ctx := context.Background()

	// 创建一个会在启动时失败但在停止时也会失败的插件
	badPlugin := NewMockPlugin("double-bad-plugin")
	badPlugin.SetStartError(errors.New("start failure"))
	badPlugin.SetStopError(errors.New("stop failure"))

	// 注册插件
	err := pm.RegisterPlugin(badPlugin)
	assert.NoError(t, err)

	// 启动应该失败
	err = pm.StartAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start failure")

	// 验证管理器没有运行
	assert.False(t, pm.running)

	// 停止未运行的管理器应该失败
	err = pm.StopAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// BenchmarkPluginManager_Lifecycle 性能测试：完整生命周期
func BenchmarkPluginManager_Lifecycle(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm := NewPluginManager()

		// 注册几个插件
		for j := 0; j < 5; j++ {
			plugin := NewMockPlugin(fmt.Sprintf("lifecycle-plugin-%d-%d", i, j))
			pm.RegisterPlugin(plugin)
		}

		// 启动和停止
		pm.StartAll(ctx)
		pm.StopAll()
	}
}

// TestPluginManager_ContextCancellation 测试上下文取消功能
func TestPluginManager_ContextCancellation(t *testing.T) {
	pm := NewPluginManager()

	// 创建一个会响应取消的插件
	plugin := NewMockPlugin("cancel-test-plugin")
	pm.RegisterPlugin(plugin)

	ctx, cancel := context.WithCancel(context.Background())

	// 启动插件
	err := pm.StartAll(ctx)
	require.NoError(t, err)

	// 取消上下文
	cancel()

	// 等待一小段时间让插件响应取消
	time.Sleep(100 * time.Millisecond)

	// 停止管理器
	err = pm.StopAll()
	assert.NoError(t, err)
}

// BenchmarkPluginManager_RegisterPlugin 性能测试：注册插件
func BenchmarkPluginManager_RegisterPlugin(b *testing.B) {
	pm := NewPluginManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin := NewMockPlugin(fmt.Sprintf("bench-plugin-%d", i))
		pm.RegisterPlugin(plugin)
	}
}

// BenchmarkPluginManager_GetPlugin 性能测试：获取插件
func BenchmarkPluginManager_GetPlugin(b *testing.B) {
	pm := NewPluginManager()

	// 预先注册插件
	for i := 0; i < 1000; i++ {
		plugin := NewMockPlugin(fmt.Sprintf("bench-get-plugin-%d", i))
		pm.RegisterPlugin(plugin)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pluginName := fmt.Sprintf("bench-get-plugin-%d", i%1000)
		pm.GetPlugin(pluginName)
	}
}

// TestPluginManager_GetSyncEvent 测试获取单个插件的同步事件
func TestPluginManager_GetSyncEvent(t *testing.T) {
	pm := NewPluginManager()
	plugin := NewMockPlugin("sync-test-plugin")

	err := pm.RegisterPlugin(plugin)
	require.NoError(t, err)

	ctx := context.Background()
	err = pm.StartAll(ctx)
	require.NoError(t, err)
	defer pm.StopAll()

	// 发送一个测试事件
	testEvent := "test-event-data"
	go func() {
		time.Sleep(100 * time.Millisecond)
		plugin.SendEvent(testEvent)
	}()

	// 测试同步获取事件
	event, err := pm.GetSyncEvent("sync-test-plugin", "request", time.Second)
	require.NoError(t, err)
	assert.Equal(t, testEvent, event)

	// 测试超时情况
	_, err = pm.GetSyncEvent("sync-test-plugin", "request", 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")

	// 测试不存在的插件
	_, err = pm.GetSyncEvent("non-existent-plugin", "request", time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestPluginManager_GetSyncEvents 测试批量获取同步事件
func TestPluginManager_GetSyncEvents(t *testing.T) {
	pm := NewPluginManager()
	plugin1 := NewMockPlugin("sync-plugin-1")
	plugin2 := NewMockPlugin("sync-plugin-2")

	err := pm.RegisterPlugin(plugin1)
	require.NoError(t, err)
	err = pm.RegisterPlugin(plugin2)
	require.NoError(t, err)

	ctx := context.Background()
	err = pm.StartAll(ctx)
	require.NoError(t, err)
	defer pm.StopAll()

	// 发送测试事件
	go func() {
		time.Sleep(100 * time.Millisecond)
		plugin1.SendEvent("event-1")
		plugin2.SendEvent("event-2")
	}()

	// 测试批量同步获取事件
	results, err := pm.GetSyncEvents([]string{"sync-plugin-1", "sync-plugin-2"}, "request", time.Second)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "event-1", results["sync-plugin-1"])
	assert.Equal(t, "event-2", results["sync-plugin-2"])

	// 测试空插件列表
	results, err = pm.GetSyncEvents([]string{}, "request", time.Second)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

// TestPluginManager_GetAllSyncEvents 测试获取所有插件的同步事件
func TestPluginManager_GetAllSyncEvents(t *testing.T) {
	pm := NewPluginManager()
	plugin1 := NewMockPlugin("all-sync-plugin-1")
	plugin2 := NewMockPlugin("all-sync-plugin-2")

	err := pm.RegisterPlugin(plugin1)
	require.NoError(t, err)
	err = pm.RegisterPlugin(plugin2)
	require.NoError(t, err)

	ctx := context.Background()
	err = pm.StartAll(ctx)
	require.NoError(t, err)
	defer pm.StopAll()

	// 发送测试事件
	go func() {
		time.Sleep(100 * time.Millisecond)
		plugin1.SendEvent("all-event-1")
		plugin2.SendEvent("all-event-2")
	}()

	// 测试获取所有插件的同步事件
	results, err := pm.GetAllSyncEvents("request", time.Second)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "all-event-1", results["all-sync-plugin-1"])
	assert.Equal(t, "all-event-2", results["all-sync-plugin-2"])
}

// TestPluginManager_SupportsSyncEvents 测试检查插件是否支持同步事件
func TestPluginManager_SupportsSyncEvents(t *testing.T) {
	pm := NewPluginManager()
	plugin := NewMockPlugin("support-check-plugin")

	err := pm.RegisterPlugin(plugin)
	require.NoError(t, err)

	// MockPlugin 应该支持同步事件
	assert.True(t, pm.SupportsSyncEvents("support-check-plugin"))

	// 不存在的插件应该返回 false
	assert.False(t, pm.SupportsSyncEvents("non-existent-plugin"))
}

// BenchmarkGetSyncEvent 同步事件获取的性能测试
func BenchmarkGetSyncEvent(b *testing.B) {
	pm := NewPluginManager()
	plugin := NewMockPlugin("bench-sync-plugin")
	pm.RegisterPlugin(plugin)

	ctx := context.Background()
	pm.StartAll(ctx)
	defer pm.StopAll()

	// 预先发送事件到通道
	for i := 0; i < b.N; i++ {
		plugin.SendEvent(fmt.Sprintf("event-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.GetSyncEvent("bench-sync-plugin", "request", time.Second)
	}
}

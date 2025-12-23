package pluginmanager

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type CollectorPlugin interface {
	// 启动插件
	Start(ctx context.Context) error
	// 停止插件
	Stop(ctx context.Context) error
	// Events 返回插件的事件
	Events(ctx context.Context, value any) <-chan any
	// SyncEvents 同步获取插件事件，阻塞直到有事件或超时
	SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error)
	// 获取插件名称（用于识别和管理）
	Name() string
	// 检查插件是否正在运行
	IsRunning() bool
}

type PluginManager struct {
	plugins        map[string]CollectorPlugin
	mu             sync.RWMutex
	running        bool
	ctx            context.Context
	cancel         context.CancelFunc
	syncEventChans map[string]chan struct{}
	syncMu         sync.Mutex
}

// NewPluginManager 工厂函数
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:        make(map[string]CollectorPlugin),
		syncEventChans: make(map[string]chan struct{}),
		running:        false,
	}
}

// - 注册插件
func (pm *PluginManager) RegisterPlugin(plugin CollectorPlugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin with name '%s' already registered", plugin.Name())
	}

	pm.plugins[plugin.Name()] = plugin

	// 如果管理器已经在运行，自动启动新注册的插件
	if pm.running && pm.ctx != nil {
		if err := plugin.Start(pm.ctx); err != nil {
			delete(pm.plugins, plugin.Name())
			return fmt.Errorf("failed to start plugin '%s': %v", plugin.Name(), err)
		}
	}

	return nil
}

// - 注销插件
func (pm *PluginManager) UnregisterPlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin with name '%s' not found", name)
	}

	// 如果插件正在运行，先停止它
	if plugin.IsRunning() && pm.ctx != nil {
		if err := plugin.Stop(pm.ctx); err != nil {
			return fmt.Errorf("failed to stop plugin '%s': %v", name, err)
		}
	}

	delete(pm.plugins, name)
	return nil
}

// - 启动所有插件
func (pm *PluginManager) StartAll(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.running {
		return fmt.Errorf("plugin manager is already running")
	}

	pm.ctx, pm.cancel = context.WithCancel(ctx)
	pm.running = true

	// 启动所有已注册的插件
	for name, plugin := range pm.plugins {
		if err := plugin.Start(pm.ctx); err != nil {
			// 如果某个插件启动失败，停止所有已启动的插件
			for _, p := range pm.plugins {
				if p.IsRunning() {
					p.Stop(pm.ctx)
				}
			}
			pm.running = false
			return fmt.Errorf("failed to start plugin '%s': %v", name, err)
		}
	}

	return nil
}

// - 停止所有插件
func (pm *PluginManager) StopAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return fmt.Errorf("plugin manager is not running")
	}

	var lastErr error
	for name, plugin := range pm.plugins {
		if plugin.IsRunning() {
			if err := plugin.Stop(pm.ctx); err != nil {
				lastErr = fmt.Errorf("failed to stop plugin '%s': %v", name, err)
			}
		}
	}

	if pm.cancel != nil {
		pm.cancel()
	}

	pm.running = false
	pm.ctx = nil
	pm.cancel = nil

	return lastErr
}

// - 获取插件列表
func (pm *PluginManager) GetPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// 不安全的获取插件方法
func (pm *PluginManager) UnsafeGetPlugin(name string) (CollectorPlugin, bool) {
	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// - 获取特定插件
func (pm *PluginManager) GetPlugin(name string) (CollectorPlugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// - 同步获取单个插件事件
func (pm *PluginManager) GetSyncEvent(name string, value any, timeout time.Duration) (any, error) {
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin with name '%s' not found", name)
	}

	// 检查插件是否支持同步事件
	if syncPlugin, ok := plugin.(interface {
		SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error)
	}); ok {
		ctx := pm.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		return syncPlugin.SyncEvents(ctx, value, timeout)
	}

	// 如果插件不支持同步事件，使用现有的事件通道
	ctx := pm.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	select {
	case event := <-plugin.Events(ctx, value):
		return event, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for event from plugin '%s'", name)
	}
}

// - 批量同步获取多个插件事件
func (pm *PluginManager) GetSyncEvents(pluginNames []string, value any, timeout time.Duration) (map[string]any, error) {
	if len(pluginNames) == 0 {
		return make(map[string]any), nil
	}

	ctx := pm.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	results := make(map[string]any)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, name := range pluginNames {
		wg.Add(1)
		go func(pluginName string) {
			defer wg.Done()

			event, err := pm.GetSyncEvent(pluginName, value, timeout)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			results[pluginName] = event
		}(name)
	}

	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}

// - 获取所有插件的同步事件
func (pm *PluginManager) GetAllSyncEvents(value any, timeout time.Duration) (map[string]any, error) {
	pm.mu.RLock()
	pluginNames := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		pluginNames = append(pluginNames, name)
	}
	pm.mu.RUnlock()

	return pm.GetSyncEvents(pluginNames, value, timeout)
}

// - 检查插件是否支持同步事件
func (pm *PluginManager) SupportsSyncEvents(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return false
	}

	_, ok := plugin.(interface {
		SyncEvents(ctx context.Context, value any, timeout time.Duration) (any, error)
	})

	return ok
}

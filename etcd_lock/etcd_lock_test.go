//go:build ignore

package etcd_lock

import (
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func setupEtcdClient(t *testing.T) *clientv3.Client {
	config := clientv3.Config{
		Endpoints: []string{"172.16.5.7:2379"},
		Username:  "root",
		Password:  "Wf3DpvKbiZ+R1rZ2",
		// 设置较短的超时时间用于测试
		DialTimeout: 2 * time.Second,
	}

	client, err := clientv3.New(config)
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	return client
}

func TestEtcdLock(t *testing.T) {
	client := setupEtcdClient(t)
	defer client.Close()

	t.Run("acquire lock successfully", func(t *testing.T) {
		lock := NewEtcdLock(client, "/test/lock1")
		err := lock.Lock(5) // 5秒的TTL
		if err != nil {
			t.Errorf("Failed to acquire lock: %v", err)
		}
		defer lock.Unlock()
	})

	t.Run("fail to acquire locked key", func(t *testing.T) {
		// 第一个锁
		lock1 := NewEtcdLock(client, "/test/lock2")
		err := lock1.Lock(5)
		if err != nil {
			t.Errorf("Failed to acquire first lock: %v", err)
		}
		defer lock1.Unlock()

		// 尝试获取同一个key的第二个锁
		lock2 := NewEtcdLock(client, "/test/lock2")
		err = lock2.Lock(5)
		if err != LockAcquiredError {
			t.Errorf("Expected LockAcquiredError, got: %v", err)
		}
	})

	t.Run("release lock successfully", func(t *testing.T) {
		lock1 := NewEtcdLock(client, "/test/lock3")
		err := lock1.Lock(5)
		if err != nil {
			t.Errorf("Failed to acquire lock: %v", err)
		}

		// 释放锁
		err = lock1.Unlock()
		if err != nil {
			t.Errorf("Failed to release lock: %v", err)
		}

		// 验证锁已释放，通过尝试重新获取
		lock2 := NewEtcdLock(client, "/test/lock3")
		err = lock2.Lock(5)
		if err != nil {
			t.Errorf("Failed to acquire lock after release: %v", err)
		}
		defer lock2.Unlock()
	})

	t.Run("auto renewal works", func(t *testing.T) {
		lock := NewEtcdLock(client, "/test/lock4")
		err := lock.Lock(5)
		if err != nil {
			t.Errorf("Failed to acquire lock: %v", err)
		}
		defer lock.Unlock()

		// 等待超过TTL时间，如果自动续约正常工作，锁应该仍然存在
		time.Sleep(6 * time.Second)

		// 尝试获取同一个key的锁，应该失败
		lock2 := NewEtcdLock(client, "/test/lock4")
		err = lock2.Lock(5)
		if err != LockAcquiredError {
			t.Errorf("Expected LockAcquiredError after TTL, got: %v", err)
		}
	})
}

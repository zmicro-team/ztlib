package etcd_lock

import (
	"context"
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// LockAcquired 获取锁失败
var LockAcquiredError = errors.New("get lock failed")

type EtcdLock struct {
	client *clientv3.Client
	// lease   clientv3.Lease   // 租约
	leaseID clientv3.LeaseID // 租约ID
	cancel  context.CancelFunc
	key     string
	ctx     context.Context
}

func NewEtcdLock(client *clientv3.Client, key string) *EtcdLock {
	return &EtcdLock{
		client: client,
		key:    key,
		ctx:    context.Background(),
	}
}

// Lock 尝试获取锁
func (l *EtcdLock) Lock(ttl int64) error {
	// 创建租约
	resp, err := l.client.Grant(l.ctx, ttl)
	if err != nil {
		return err
	}
	l.leaseID = resp.ID

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(l.ctx)
	l.cancel = cancel

	// 自动续约
	keepAlive, err := l.client.KeepAlive(ctx, l.leaseID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			// 自动续约
			case _, ok := <-keepAlive:
				if !ok {
					return
				}
			// 上下文被取消
			case <-ctx.Done():
				return
			}
		}
	}()

	// 尝试获取锁
	txn := l.client.Txn(l.ctx).
		If(clientv3.Compare(clientv3.CreateRevision(l.key), "=", 0)).
		Then(clientv3.OpPut(l.key, "", clientv3.WithLease(l.leaseID))).
		Else()

	txnResp, err := txn.Commit()
	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		return LockAcquiredError
	}

	return nil
}

// Unlock 释放锁
func (l *EtcdLock) Unlock() error {
	if l.cancel != nil {
		l.cancel()
	}

	if _, err := l.client.Revoke(l.ctx, l.leaseID); err != nil {
		return err
	}

	return nil
}

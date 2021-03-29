// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

const (
	TimeoutSecond   = 5
	DefaultLeaseTTL = 5
)

type ServiceSinker interface {
	NodeInfo() *protocol.NodeInfo
	AddDependency(*protocol.NodeInfo)
	DelDependency(bool, fatchoy.NodeID)
}

// 基于ETCD的服务发现
type EtcdDiscovery struct {
	done      chan struct{}
	wg        sync.WaitGroup
	closing   int32              //
	endpoints []string           //
	keySpace  string             //
	leaseTTL  int                // lease TTL seconds
	ctx       context.Context    // master context
	cancel    context.CancelFunc //
	client    *clientv3.Client   // etcd client
	leaseID   clientv3.LeaseID   // etcd lease ID
	sink      ServiceSinker      //
}

func NewEtcdDiscovery(env *fatchoy.Environ, sink ServiceSinker) *EtcdDiscovery {
	d := &EtcdDiscovery{
		endpoints: strings.Split(env.EtcdAddr, ","),
		keySpace:  fmt.Sprintf("%s/service", env.EtcdKeyspace),
		leaseTTL:  int(env.EtcdLeaseTtl),
		done:      make(chan struct{}),
		ctx:       context.Background(),
		sink:      sink,
	}
	if d.leaseTTL <= 0 {
		d.leaseTTL = DefaultLeaseTTL
	}
	return d
}

func (d *EtcdDiscovery) Start() error {
	if err := d.makeClient(); err != nil {
		return err
	}
	if err := d.listService(); err != nil {
		return err
	}

	d.ctx, d.cancel = context.WithCancel(context.Background())
	watchCh := d.client.Watch(clientv3.WithRequireLeader(d.ctx), d.keySpace, clientv3.WithPrefix())
	leaseCh, err := d.register(d.ctx)
	if err != nil {
		return err
	}
	go d.serve(leaseCh, watchCh)

	return nil
}

func (d *EtcdDiscovery) makeClient() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   d.endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	d.client = cli
	d.ctx = context.Background()
	return nil
}

func (d *EtcdDiscovery) listService() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*TimeoutSecond)
	defer cancel()
	resp, err := d.client.Get(ctx, d.keySpace, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if resp != nil {
		for _, v := range resp.Kvs {
			d.addDependency(v.Key, v.Value)
		}
	}
	return nil
}

func (d *EtcdDiscovery) register(leaseCtx context.Context) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*TimeoutSecond)
	defer cancel()
	info := d.sink.NodeInfo()
	key := fmt.Sprintf("%s/%s", d.keySpace, fatchoy.NodeID(info.Node).String())
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if resp != nil && len(resp.Kvs) > 0 {
		return nil, errors.Errorf("duplicate registration of %s", key)
	}
	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*TimeoutSecond)
	defer cancel()
	lease, err := d.client.Grant(ctx, int64(d.leaseTTL))
	if err != nil {
		return nil, err
	}
	d.leaseID = lease.ID

	ctx, cancel = context.WithTimeout(ctx, time.Second*TimeoutSecond)
	defer cancel()
	if _, err := d.client.Put(ctx, key, string(data), clientv3.WithLease(lease.ID)); err != nil {
		return nil, err
	}

	return d.client.KeepAlive(leaseCtx, lease.ID)
}

func (d *EtcdDiscovery) reconnect() (<-chan *clientv3.LeaseKeepAliveResponse, clientv3.WatchChan, error) {
	if err := d.makeClient(); err != nil {
		return nil, nil, err
	}
	if err := d.listService(); err != nil {
		return nil, nil, err
	}
	if d.cancel != nil {
		d.cancel()
	}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	watchCh := d.client.Watch(clientv3.WithRequireLeader(d.ctx), d.keySpace, clientv3.WithPrefix())
	leaseCh, err := d.register(d.ctx)
	if err != nil {
		return nil, nil, err
	}
	return leaseCh, watchCh, nil
}

func (d *EtcdDiscovery) serve(lch <-chan *clientv3.LeaseKeepAliveResponse, wch clientv3.WatchChan) {
	defer d.wg.Done()
	defer d.revoke()

	ticker := time.NewTicker(time.Millisecond * 1500)
	defer ticker.Stop()

	for {
		select {
		case lease, ok := <-lch:
			if !ok || lease == nil {
				log.Errorf("lost connection to etcd server [%s]", d.endpoints)
				lch = nil
				wch = nil
				atomic.StoreInt32(&d.closing, 1)
			}

		case rsp := <-wch:
			for _, ev := range rsp.Events {
				switch ev.Type {
				case 0: // mvccpb.PUT
					d.addDependency(ev.Kv.Key, ev.Kv.Value)
				case 1: // mvccpb.DELETE
					d.delDependency(ev.Kv.Key, ev.Kv.Value)
				}
			}

		case <-ticker.C:
			if atomic.LoadInt32(&d.closing) > 0 {
				d.sink.DelDependency(true, 0)
				d.client.Close()
				if leaseCh, watchCh, err := d.reconnect(); err != nil {
					log.Errorf("reconnect etcd: %v", err)
				} else {
					lch, wch = leaseCh, watchCh
					atomic.StoreInt32(&d.closing, 0)
				}
			}

		case <-d.done:
			return
		}
	}
}

func (d *EtcdDiscovery) revoke() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	if _, err := d.client.Revoke(ctx, d.leaseID); err != nil {
		log.Errorf("Delete alive key: %v", err)
	}
	d.leaseID = 0
}

func (d *EtcdDiscovery) addDependency(key, value []byte) {
	//log.Infof("add dependency: %s: %s", key, value)
	var info protocol.NodeInfo
	if err := json.Unmarshal(value, &info); err != nil {
		log.Errorf("marshal %s[%s]: %v", key, value, err)
	} else {
		d.sink.AddDependency(&info)
	}
}

func (d *EtcdDiscovery) delDependency(key, value []byte) {
	//log.Infof("del dependency: %s: %s", key, value)
	i := bytes.LastIndexByte(key, '/')
	if i <= 0 {
		log.Errorf("cannot index node id of key: %s", key)
		return
	}
	s := string(key[i+1:])
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Errorf("cannot parse node id of key: %s, %v", key, err)
		return
	}
	d.sink.DelDependency(false, fatchoy.NodeID(n))
}

func (d *EtcdDiscovery) Close() {
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	close(d.done)
	d.wg.Wait()
	d.client.Close()
}

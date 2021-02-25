// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

import (
	"log"
	"sort"
	"sync"
)

var (
	regGuard        sync.RWMutex
	serviceRegistry = make(map[string]Service)
	serviceIdMap    = make(map[uint8]string)
)

// 注册服务
func Register(service Service) {
	regGuard.Lock()
	defer regGuard.Unlock()
	var name = service.Name()
	var id = service.ID()
	if _, dup := serviceRegistry[name]; dup {
		log.Panicf("duplicate registration of service %x", name)
	}
	if _, dup := serviceIdMap[id]; dup {
		log.Panicf("duplicate ID of service %x", id)
	}
	serviceRegistry[name] = service
	serviceIdMap[id] = name
}

// 根据服务ID获取Service对象
func GetServiceByID(srvType uint8) Service {
	var v Service
	regGuard.RLock()
	if name, ok := serviceIdMap[srvType]; ok {
		v = serviceRegistry[name]
	}
	regGuard.RUnlock()
	return v
}

// 根据服务类型名获取服务类型
func GetServiceTypeByName(name string) uint8 {
	var srvType uint8
	regGuard.RLock()
	if srv, found := serviceRegistry[name]; found {
		srvType = srv.ID()
	}
	regGuard.RUnlock()
	return srvType
}

// 根据名称获取Service对象
func GetServiceByName(name string) Service {
	regGuard.RLock()
	v := serviceRegistry[name]
	regGuard.RUnlock()
	return v
}

// 所有服务类型名
func GetServiceNames() []string {
	regGuard.RLock()
	var names = make([]string, 0, len(serviceRegistry))
	for s, _ := range serviceRegistry {
		names = append(names, s)
	}
	regGuard.RUnlock()
	sort.Strings(names)
	return names
}

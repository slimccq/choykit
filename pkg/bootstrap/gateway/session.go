// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"devpkg.work/choykit/pkg/fatchoy"
)

//	client node layout:
//  -----------------------------------
// 	| type |  gid  |   session id     |
// 	-----------------------------------
//  32    31      24                  0
//
// 最高位为1，代表是client session
// 中间7位为同一个区服内gate的分组号
// 低24位是客户端session号
// 这样保证同一个区服的gate能分配唯一的node号

const (
	MaxSessionID = (1 << 24) - 1
)

func SessionToNode(node fatchoy.NodeID, sid uint32) fatchoy.NodeID {
	var gid = uint32(node.Instance()) & 0x7F
	return fatchoy.NodeTypeClient | fatchoy.NodeID(gid<<24) | fatchoy.NodeID(sid)
}

func NodeToSession(node fatchoy.NodeID) uint32 {
	return uint32(node) & MaxSessionID
}

// 分配下一个session序列号
func (g *Service) nextSession() fatchoy.NodeID {
	for {
		g.nextSid++
		if g.nextSid >= MaxSessionID {
			g.nextSid = 0
		}
		var node = SessionToNode(g.NodeID(), g.nextSid)
		if g.sessions.Get(node) != nil {
			continue
		}
		return node
	}
}

// 玩家会话的簿记信息
type SessionUserData struct {
	userid int64
	session uint32
}

func NewSessionUserData() *SessionUserData {
	return &SessionUserData{}
}

func GetSessionUData(session fatchoy.Endpoint) *SessionUserData {
	if v, ok := session.UserData().(*SessionUserData); ok {
		return v
	}
	return nil
}

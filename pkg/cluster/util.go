// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"devpkg.work/choykit/pkg/fatchoy"
)

// 注册签名
func SignAccessToken(node fatchoy.NodeID, gameId, key string) string {
	var buf bytes.Buffer
	buf.WriteString(node.String())
	buf.WriteString(gameId)
	h := hmac.New(sha256.New, []byte(key))
	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}

func DependencyServiceTypes(serviceDependency string) ([]uint8, error) {
	var services []uint8
	for _, name := range strings.Split(serviceDependency, ",") {
		if srvType := fatchoy.GetServiceTypeByName(name); srvType > 0 {
			services = append(services, srvType)
		} else {
			return nil, fmt.Errorf("unrecognized dependency type %s", name)
		}
	}
	return services, nil
}

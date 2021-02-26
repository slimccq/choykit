// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

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

// Copyright © 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package cluster

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"devpkg.work/choykit/pkg"
)

// 注册签名
func SignAccessToken(node choykit.NodeID, gameId, key string) string {
	var buf bytes.Buffer
	buf.WriteString(node.String())
	buf.WriteString(gameId)
	h := hmac.New(sha256.New, []byte(key))
	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}

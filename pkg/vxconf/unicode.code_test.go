// Copyright © 2020 Hedzr Yeh.

package vxconf

import (
	"testing"
)

func TestUnescape(t *testing.T) {
	str := `id: fwd-http-mock-1
type: mock
match: /api/mock-1
to: {}
mocks:
- methods: [GET, PUT, DELETE, POST, HEAD, OPTIONS, TRACE, CONNECT]
  object: {data: [{age: 13, name: jasper, sex: male}], status: ok}
  text: "\u591A\u884C\u6587\u672C\n"
lb: {}
auth: {}
`
	str = UnescapeUnicode([]byte(str))
	t.Logf("Result: %s", str)

	str = `id: fwd-ms-consulapi
type: ms
match: /api/test1
to:
  ms: consulapi
  context: /v1
lb: {}
downgrade-to-http1: true
reverse-rewrite: true
desc: "这是consulapi服务的一个包装转换，向外部提供核心consul  matrix \u7684HTTP API\u63A5\u53E3\u3002\u8BF7\u786E\u4FDD\u751F\u4EA7\u73AF\u5883\u4E2D\u5DF2\u7ECF\u5173\u95ED\u4E86\u6B64\u8F6C\u53D1\u5668\u3002"
auth: {}
`
	str = UnescapeUnicode([]byte(str))
	str = UnescapeUnicode([]byte(str))
	t.Logf("Result: %s", str)
}

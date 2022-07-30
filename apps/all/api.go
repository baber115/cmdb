package all

import (
	// 注册所有HTTP服务模块, 暴露给框架HTTP服务器加载
	_ "codeup.aliyun.com/baber/go/cmdb/apps/book/api"
	_ "codeup.aliyun.com/baber/go/cmdb/apps/host/api"
	_ "codeup.aliyun.com/baber/go/cmdb/apps/resource/api"
)

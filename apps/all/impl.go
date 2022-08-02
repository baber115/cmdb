package all

import (
	// 注册所有GRPC服务模块, 暴露给框架GRPC服务器加载, 注意 导入有先后顺序
	_ "codeup.aliyun.com/baber/go/cmdb/apps/book/impl"
	_ "codeup.aliyun.com/baber/go/cmdb/apps/host/impl"
	_ "codeup.aliyun.com/baber/go/cmdb/apps/resource/impl"
	_ "codeup.aliyun.com/baber/go/cmdb/apps/secret/impl"
)

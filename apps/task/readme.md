# 云资源任务管理

+ secret
+ provider
+ host service

例子：
/task/
```
# 资源同步任务
type : "sync"
# secret,比如腾讯云的secret
secret_id: "xxx"
# operater 按照资源划分，操作主机
resource_type: "host"
# 操作某个地区的资源，云商不能操作跨地区的资源
region: "shanghai"
```

任务的状态
+ 状态：running
+ 开启时间
+ 介绍时间
+ 执行日志
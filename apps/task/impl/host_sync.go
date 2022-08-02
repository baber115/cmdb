package impl

import (
	"context"
	"fmt"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	"codeup.aliyun.com/baber/go/cmdb/apps/secret"
	"codeup.aliyun.com/baber/go/cmdb/apps/task"
	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/connectivity"
	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/cvm"
)

type syncHostRequest struct {
	secret *secret.Secret
	task   *task.Task
}

func newSyncHostRequest(secret *secret.Secret, task *task.Task) *syncHostRequest {
	return &syncHostRequest{
		secret: secret,
		task:   task,
	}
}

func (i *impl) syncHost(ctx context.Context, req *syncHostRequest) {
	// 处理任务状态
	// go runtine里面一定要捕获异常，否则会程序崩掉
	// go recover只能捕获当前Goruntine的panic
	defer func() {
		if err := recover(); err != nil {
			req.task.Failed(fmt.Sprintf("pannic, %v", err))
		} else {
			if !req.task.Status.Stage.IsIn(task.Stage_FAILED, task.Stage_WARNING) {
				req.task.Success()
			}
			if err := i.update(ctx, req.task); err != nil {
				i.log.Errorf("save task error, %s", err)
			}
		}
	}()

	// 实现主机同步
	txConn := connectivity.NewTencentCloudClient(
		req.secret.Data.ApiKey,
		req.secret.Data.ApiSecret,
		req.task.Data.Region,
	)
	cvmOp := cvm.NewCVMOperator(txConn.CvmClient())

	// 因为要同步所有资源，需要分页查询
	pagger := cvm.NewPagger(cvmOp)
	for pagger.HasNext() {
		set := host.NewHostSet()
		// 查询分页有错误 反应在Task上面
		if err := pagger.Scan(ctx, set); err != nil {
			req.task.Failed(err.Error())
			return
		}

		// 保持该页数据, 同步时间时, 记录下日志
		for index := range set.Items {
			_, err := i.host.SyncHost(ctx, set.Items[index])
			if err != nil {
				req.task.Failed(err.Error())
				continue
			}
		}
	}
}

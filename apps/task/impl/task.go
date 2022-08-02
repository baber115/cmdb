package impl

import (
	"context"
	"fmt"
	"time"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"codeup.aliyun.com/baber/go/cmdb/apps/secret"
	"codeup.aliyun.com/baber/go/cmdb/apps/task"
	"codeup.aliyun.com/baber/go/cmdb/conf"
	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/connectivity"
	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/cvm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *impl) CreateTask(ctx context.Context, req *task.CreateTaskRequst) (*task.Task, error) {
	t, err := task.CreateTask(req)
	if err != nil {
		return nil, err
	}
	// 1. 查询secret
	s, err := i.secret.DescribeSecret(ctx, secret.NewDescribeSecretRequest(req.SecretId))
	if err != nil {
		return nil, err
	}
	t.SecretDescription = s.Data.Description

	// 并解密api secret
	if err := s.Data.DecryptAPISecret(conf.C().App.EncryptKey); err != nil {
		return nil, err
	}

	// 需要把Task 标记为Running, 修改一下Task对象的状态
	t.Run()

	var taskCancel context.CancelFunc
	switch req.Type {
	case task.Type_RESOURCE_SYNC:
		// 根据secret所属的厂商, 初始化对应厂商的operator
		switch s.Data.Vendor {
		case resource.Vendor_TENCENT:
			// 操作那种资源:
			switch req.ResourceType {
			case resource.Type_HOST:
				// 只实现主机同步, 初始化腾讯cvm operator
				// NewTencentCloudClient client
				txConn := connectivity.NewTencentCloudClient(s.Data.ApiKey, s.Data.ApiSecret, req.Region)
				cvmOp := cvm.NewCVMOperator(txConn.CvmClient())

				// 因为要同步所有资源，需要分页查询
				pagger := cvm.NewPagger(cvmOp)
				for pagger.HasNext() {
					set := host.NewHostSet()
					// 查询分页有错误立即返回
					if err := pagger.Scan(ctx, set); err != nil {
						return nil, err
					}
					// 保持该页数据, 同步时间时, 记录下日志
					for index := range set.Items {
						_, err := i.host.SyncHost(ctx, set.Items[index])
						if err != nil {
							i.log.Errorf("sync host error, %s", err)
							continue
						}
					}
				}
				// 直接使用goroutine 把最耗时的逻辑
				// ctx 是不是传递Http 的ctx
				taskExecCtx, cancel := context.WithTimeout(
					context.Background(),
					time.Duration(req.Timeout)*time.Second,
				)
				taskCancel = cancel

				go i.syncHost(taskExecCtx, newSyncHostRequest(s, t))
			case resource.Type_RDS:
			case resource.Type_BILL:
			}
		case resource.Vendor_ALIYUN:

		case resource.Vendor_HUAWEI:

		case resource.Vendor_AMAZON:

		case resource.Vendor_VSPHERE:
		default:
			return nil, fmt.Errorf("unknow resource type: %s", s.Data.Vendor)
		}

		// 2. 利用secret的信息, 初始化一个operater
		// 使用operator进行资源的操作, 比如同步

		// 调用host service 把数据入库
	case task.Type_RESOURCE_RELEASE:
	default:
		return nil, fmt.Errorf("unknow task type: %s", req.Type)
	}

	if err := i.insert(ctx, t); err != nil {
		if taskCancel != nil {
			taskCancel()
		}
		return nil, err
	}

	return t, nil
}
func (i *impl) QueryTask(ctx context.Context, req *task.QueryTaskRequest) (*task.TaskSet, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTask not implemented")
}
func (i *impl) DescribeTask(ctx context.Context, req *task.DescribeTaskRequest) (*task.Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeTask not implemented")
}

package impl

import (
	"context"
	"fmt"
	"time"

	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"codeup.aliyun.com/baber/go/cmdb/apps/secret"
	"codeup.aliyun.com/baber/go/cmdb/apps/task"
	"codeup.aliyun.com/baber/go/cmdb/conf"
	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/sqlbuilder"
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
				// 直接使用goroutine 把最耗时的逻辑
				taskExecCtx, cancel := context.WithTimeout(
					context.Background(),
					time.Duration(req.Timeout)*time.Second,
				)
				taskCancel = cancel
				// 这里不能用ctx，ctx是请求的上下文，如果传给goroutine的话，请求结束时goroutine也会结束
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
	query := sqlbuilder.NewQuery(queryTaskSQL)

	sql, args := query.Limit(req.Page.ComputeOffset(), uint(req.Page.PageSize)).BuildQuery()
	i.log.Debugf("sql: %s, args: %v", sql, args)

	stmt, err := i.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, exception.NewInternalServerError("prepare query task error, %s", err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskSet := task.NewTaskSet()
	for rows.Next() {
		t := task.NewDefaultTask()

		/*
			报错：sql: Scan error on column index 4, name "secret_desc": converting driver.Value type []uint8 ("测试用例") to a int64: inv    alid syntax
			注意字段顺序，需要和数据库的保持一致
		*/
		err = rows.Scan(
			&t.Id, &t.Data.Region, &t.Data.ResourceType, &t.Data.SecretId, &t.SecretDescription, &t.Data.Timeout,
			&t.Status.Stage, &t.Status.Message, &t.Status.StartAt, &t.Status.EndAt,
			&t.Status.TotalSucceed, &t.Status.TotalFailed,
		)
		if err != nil {
			return nil, status.Errorf(codes.Unimplemented, "method QueryTask not implemented")
		}
		taskSet.Add(t)
	}

	return taskSet, nil
}

func (i *impl) DescribeTask(ctx context.Context, req *task.DescribeTaskRequest) (*task.Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeTask not implemented")
}

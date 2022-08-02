package impl_test

import (
	"context"
	"os"
	"testing"

	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"codeup.aliyun.com/baber/go/cmdb/apps/task"
	"codeup.aliyun.com/baber/go/cmdb/conf"
	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger/zap"

	// 注册所有对象，不然会报secret没有注册
	_ "codeup.aliyun.com/baber/go/cmdb/apps/all"
)

var (
	ins task.ServiceServer
)

func init() {
	if err := conf.LoadConfigFromEnv(); err != nil {
		panic(err)
	}

	zap.DevelopmentSetup()

	if err := app.InitAllApp(); err != nil {
		panic(err)
	}
	ins = app.GetGrpcApp(task.AppName).(task.ServiceServer)
}

func TestCreateTask(t *testing.T) {
	req := task.NewCreateTaskRequst()
	req.Type = task.Type_RESOURCE_SYNC
	req.Region = os.Getenv("TX_CLOUD_REGION")
	req.ResourceType = resource.Type_HOST
	req.SecretId = "cbkacn0aeig0d12sitl0"

	taskIns, err := ins.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(taskIns)
}

package impl_test

import (
	"context"
	"os"
	"testing"

	"codeup.aliyun.com/baber/go/cmdb/conf"
	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger/zap"

	_ "codeup.aliyun.com/baber/go/cmdb/apps/all"
	"codeup.aliyun.com/baber/go/cmdb/apps/secret"
)

var (
	ins        secret.ServiceServer
	encryptKey = "abc"
)

func init() {
	// 通过环境变量加载测试配置
	if err := conf.LoadConfigFromEnv(); err != nil {
		panic(err)
	}

	// 全局日志对象初始化
	zap.DevelopmentSetup()

	// 初始化所有实例
	if err := app.InitAllApp(); err != nil {
		panic(err)
	}

	ins = app.GetGrpcApp(secret.AppName).(secret.ServiceServer)
}

func TestSecretEncrypt(t *testing.T) {
	ins := secret.NewDefaultSecret()
	ins.Data.ApiSecret = "123456"
	ins.Data.EncryptAPISecret(encryptKey)
	t.Log(ins.Data.ApiSecret)
	ins.Data.DecryptAPISecret(encryptKey)
	t.Log(ins.Data.ApiSecret)
}

func TestQuerySecret(t *testing.T) {
	ss, err := ins.QuerySecret(context.Background(), secret.NewQuerySecretRequest())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
}

func TestDescribeSecret(t *testing.T) {
	ss, err := ins.DescribeSecret(context.Background(), secret.NewDescribeSecretRequest("cbjobp8aeig7360lemn0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
}

func TestCreateSecret(t *testing.T) {
	req := secret.NewCreateSecretRequest()
	req.Description = "测试用例"
	req.ApiKey = os.Getenv("TX_CLOUD_SECRET_ID")
	req.ApiSecret = os.Getenv("TX_CLOUD_SECRET_KEY")
	req.AllowRegions = []string{"*"}
	ss, err := ins.CreateSecret(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
}

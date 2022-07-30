package cvm_test

import (
	"context"
	"fmt"
	"testing"

	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/connectivity"
	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/cvm"
	"github.com/infraboard/mcube/logger/zap"
	tx_cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var (
	operator *cvm.CVMOperator
)

func init() {
	// 初始化client
	err := connectivity.LoadClientFromEnv()
	if err != nil {
		panic(err)
	}

	// 初始化日志
	zap.DevelopmentSetup()

	operator = cvm.NewCVMOperator(connectivity.C().CvmClient())
}

func TestOperator(t *testing.T) {
	request := tx_cvm.NewDescribeInstancesRequest()
	response, err := operator.Query(context.Background(), request)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
}

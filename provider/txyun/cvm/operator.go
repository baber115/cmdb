package cvm

import (
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

func NewCVMOperator(conn *cvm.Client) *CVMOperator {
	return &CVMOperator{
		client: conn,
		log:    zap.L().Named("Tx CVM"),
	}
}

type CVMOperator struct {
	client *cvm.Client
	log    logger.Logger
}

package connectivity

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/infraboard/cmdb/utils"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

var (
	client *TencentCloudClient
)

func C() *TencentCloudClient {
	if client == nil {
		panic("please load config first")
	}
	return client
}

// TencentCloudClient client for all TencentCloud service
type TencentCloudClient struct {
	Region    string `env:"TX_CLOUD_REGION"`
	SecretID  string `env:"TX_CLOUD_SECRET_ID"`
	SecretKey string `env:"TX_CLOUD_SECRET_KEY"`

	cvmConn *cvm.Client
	cdbConn *cdb.Client
	//redisConn *redis.Client
	//cosConn   *cos.Client
	billConn *billing.Client
}

func LoadClientFromEnv() error {
	client = &TencentCloudClient{}
	if err := env.Parse(client); err != nil {
		return err
	}
	fmt.Println(client)
	return nil
}

// NewTencentCloudClient client
func NewTencentCloudClient(credentialID, credentialKey, region string) *TencentCloudClient {
	return &TencentCloudClient{
		Region:    region,
		SecretID:  credentialID,
		SecretKey: credentialKey,
	}
}

// UseCvmClient cvm
func (me *TencentCloudClient) CvmClient() *cvm.Client {
	if me.cvmConn != nil {
		return me.cvmConn
	}

	credential := common.NewCredential(
		me.SecretID,
		me.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300
	cpf.Language = "en-US"

	cvmConn, _ := cvm.NewClient(credential, me.Region, cpf)
	me.cvmConn = cvmConn
	return me.cvmConn
}

// 获取客户端账号ID
func (me *TencentCloudClient) Account() (string, error) {
	credential := common.NewCredential(
		me.SecretID,
		me.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300
	cpf.Language = "en-US"

	stsConn, _ := sts.NewClient(credential, me.Region, cpf)

	req := sts.NewGetCallerIdentityRequest()

	resp, err := stsConn.GetCallerIdentity(req)
	if err != nil {
		return "", fmt.Errorf("unable to initialize the STS client: %#v", err)
	}

	// 避免空指针
	return utils.PtrStrV(resp.Response.AccountId), nil
}

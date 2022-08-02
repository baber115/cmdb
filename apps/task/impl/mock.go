package impl

// 开发过程中，可能secret是其他同事负责，我们为了模块之前解耦，可以先定义一个secretMock，完成task的开发，等secret模块完成后替换成secret

import (
	"context"

	"codeup.aliyun.com/baber/go/cmdb/apps/secret"
)

type secretMock struct {
	secret.UnimplementedServiceServer
}

func (m *secretMock) CreateSecret(context.Context, *secret.CreateSecretRequest) (*secret.Secret, error) {
	return nil, nil
}

func (m *secretMock) QuerySecret(context.Context, *secret.QuerySecretRequest) (*secret.SecretSet, error) {
	return nil, nil
}

func (m *secretMock) DescribeSecret(context.Context, *secret.DescribeSecretRequest) (
	*secret.Secret, error) {
	ins := secret.NewDefaultSecret()
	ins.Data.ApiKey = ""
	ins.Data.ApiSecret = ""
	return ins, nil
}

func (m *secretMock) DeleteSecret(context.Context, *secret.DeleteSecretRequest) (*secret.Secret, error) {
	return nil, nil
}

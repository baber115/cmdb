package connectivity_test

import (
	"fmt"
	"testing"

	"codeup.aliyun.com/baber/go/cmdb/provider/txyun/connectivity"
	"github.com/stretchr/testify/assert"
)

func TestTencentCloudClient(t *testing.T) {
	should := assert.New(t)

	err := connectivity.LoadClientFromEnv()
	if should.NoError(err) {
		c := connectivity.C()
		fmt.Println(c.Account())
	}
}

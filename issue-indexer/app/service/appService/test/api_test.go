package test

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"issue-indexer/app/config"
	"issue-indexer/app/service/appService"
	"testing"
)

func TestApi(t *testing.T) {
	var (
		c         = createFakeConfig()
		_, server = createFakeService(c)
	)
	err := server.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}



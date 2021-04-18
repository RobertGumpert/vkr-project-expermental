package test

import (
	"testing"
)

func TestFlow(t *testing.T) {
	var (
		c         = createFakeConfig()
		_, server = createFakeTaskService(c)
	)
	err := server.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}

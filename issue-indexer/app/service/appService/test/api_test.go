package test

import "testing"

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

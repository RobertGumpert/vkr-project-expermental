package localStorage

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"strconv"
	"testing"
)

func getStorage() (*Storage, func(s *Storage)) {
	storage, err := NewStorage(
		NewFileProvider(
			"bagwords",
			3,
			ToStringString,
			ToStringFloat64Vector,
		),
	)
	if err != nil {
		runtimeinfo.LogError(err)
		if err := Destructor(storage); err != nil {
			runtimeinfo.LogError(err)
		}
		runtimeinfo.LogFatal(err)
	}
	return storage, func(s *Storage) {
		if err := Destructor(storage); err != nil {
			runtimeinfo.LogFatal(err)
		}
		runtimeinfo.LogInfo("OK")
	}
}

func TestWritingFlow(t *testing.T) {
	storage, destructor := getStorage()
	defer destructor(storage)

	keys := make([]string, 0)

	for i := 0; i<10; i++ {
		inc := strconv.Itoa(i)
		key := "Key"+inc
		data := []float64{
			1.2*float64(i), 3.4*float64(i),
		}
		err := storage.Write(
			"bagwords",
			key,
			data,
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
		keys = append(keys, key)
	}

	for _, key := range keys {
		data, err := storage.Read(
			"bagwords",
			key,
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		} else {
			runtimeinfo.LogInfo(data)
		}
	}


}

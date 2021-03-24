package localStorage

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"strconv"
	"testing"
)

var storage *Storage

func init() {
	s, _ := getStorage()
	storage = s
	// defer destructor(storage)
}

func getStorage() (*Storage, func(s *Storage)) {
	//
	storage, err := NewStorage(
		NewFileProvider(
			"bagwords",
			3,
			ToStringString,
			ToStringFloat64Vector,
			FromStringToString,
			FromStringToFloat64Vector,
		),
		NewFileProvider(
			"bench",
			3,
			ToStringString,
			ToStringFloat64Vector,
			FromStringToString,
			FromStringToFloat64Vector,
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

func TestWritingReadUpdateFlow(t *testing.T) {

	keys := make([]string, 0)
	runtimeinfo.LogInfo("START WRITE................................................")
	for i := 0; i < 10; i++ {
		inc := strconv.Itoa(i)
		key := "Key" + inc
		data := []float64{
			1.2 * float64(i), 3.4 * float64(i),
		}
		err := storage.WriteNew(
			"bagwords",
			key,
			data,
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
		keys = append(keys, key)
	}
	runtimeinfo.LogInfo("FINISH WRITE................................................")
	runtimeinfo.LogInfo("START READ................................................")
	for _, key := range keys {
		id, data, err := storage.Read(
			"bagwords",
			key,
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		} else {
			runtimeinfo.LogInfo("READ ID ", id, " WITH DATA ", data)
		}
	}
	runtimeinfo.LogInfo("FINISH READ................................................")
	runtimeinfo.LogInfo("START UPDATE................................................")
	for _, key := range keys {
		err := storage.Update(
			"bagwords",
			key,
			[]float64{
				1.2 * float64(-1), 3.4 * float64(-1),
			},
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
	}
	runtimeinfo.LogInfo("FINISH UPDATE................................................")
	runtimeinfo.LogInfo("START READ UPDATE................................................")
	for _, key := range keys {
		id, data, err := storage.Read(
			"bagwords",
			key,
		)
		if err != nil {
			runtimeinfo.LogFatal(err)
		} else {
			runtimeinfo.LogInfo("READ ID ", id, " WITH DATA ", data)
		}
	}
	runtimeinfo.LogInfo("FINISH READ UPDATE................................................")
}

func TestRewriteFlow(t *testing.T) {
	var (
		count        = 10
		id           = make([]string, 0)
		data         = make([][]float64, count+1)
		createIdList = func() {
			for i := 0; i < count; i++ {
				inc := strconv.Itoa(i)
				key := "Key" + inc
				id = append(id, key)
			}
		}
		createDataList = func() {
			for i := 0; i < count; i++ {
				vector := make([]float64, count)
				for j := 0; j < count; j++ {
					vector[j] = float64(j * i)
				}
				data[i] = vector
			}
		}
		update = func() {
			for i := 0; i < count; i++ {
				for j := 0; j < count; j++ {
					data[i][j] = data[i][j] * float64(-1)
				}
				data[i] = append(data[i], float64(0))
			}
			vector := make([]float64, count+1)
			for i := 0; i < count+1; i++ {
				vector[i] = float64(0)
			}
			data[count] = vector
			id = append(id, "Key"+strconv.Itoa(count))
		}
		convertDataToSliceInterface = func() []interface{} {
			d := make([]interface{}, 0)
			for i := 0; i < len(data); i++ {
				d = append(d, data[i])
			}
			return d
		}
		convertIdToSliceInterface = func() []interface{} {
			d := make([]interface{}, 0)
			for i := 0; i < len(id); i++ {
				d = append(d, id[i])
			}
			return d
		}
	)
	createIdList()
	createDataList()
	runtimeinfo.LogInfo("START WRITE................................................")
	for i := 0; i < len(id); i++ {
		err := storage.WriteNew(
			"bagwords",
			id[i],
			data[i],
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	runtimeinfo.LogInfo("FINISH WRITE................................................")
	//
	runtimeinfo.LogInfo("START READ................................................")
	for i := 0; i < len(id); i++ {
		id, data, err := storage.Read(
			"bagwords",
			id[i],
		)
		if err != nil {
			t.Fatal(err)
		} else {
			runtimeinfo.LogInfo("READ ID ", id, " WITH DATA ", data)
		}
	}
	runtimeinfo.LogInfo("FINISH READ................................................")
	//
	update()
	//
	runtimeinfo.LogInfo("START REWRITE................................................")
	// bagwords.txt
	err := storage.Rewrite(
		"bagwords",
		convertIdToSliceInterface(),
		convertDataToSliceInterface(),
	)
	if err != nil {
		t.Fatal(err)
	}
	runtimeinfo.LogInfo("FINISH REWRITE................................................")
	//
	runtimeinfo.LogInfo("START READ................................................")
	for i := 0; i < len(id); i++ {
		id, data, err := storage.Read(
			"bagwords",
			id[i],
		)
		if err != nil {
			t.Fatal(err)
		} else {
			runtimeinfo.LogInfo("READ ID ", id, " WITH DATA ", data)
		}
	}
	runtimeinfo.LogInfo("FINISH READ................................................")
}

func BenchmarkWriting(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inc := strconv.Itoa(i)
		key := "Key" + inc
		data := []float64{
			1.2 * float64(i), 3.4 * float64(i),
		}
		_ = storage.WriteNew(
			"bench",
			key,
			data,
		)
		//if err != nil {
		//	runtimeinfo.LogInfo("i = ",i, " err: ", err)
		//}
	}
}

func BenchmarkReading(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inc := strconv.Itoa(i)
		key := "Key" + inc
		_, _, _ = storage.Read(
			"bench",
			key,
		)
		//if err != nil {
		//	runtimeinfo.LogInfo("i = ",i, " err: ", err)
		//}
	}
}

func BenchmarkUpdate(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inc := strconv.Itoa(i)
		key := "Key" + inc
		_ = storage.Update(
			"bench",
			key,
			[]float64{
				1.2 * float64(-1), 3.4 * float64(-1),
			},
		)
		//if err != nil {
		//	runtimeinfo.LogInfo("i = ",i, " err: ", err)
		//}
	}
}

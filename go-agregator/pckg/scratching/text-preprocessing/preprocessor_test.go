package text_preprocessing

import (
	"io/ioutil"
	"log"
	"testing"
)

func ReadBytes() string {
	content, err := ioutil.ReadFile("C:/VKR/go-agregator/big.txt")
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return string(content)
}

func TestOneThreadPreprocessing(t *testing.T) {
	str:= ReadBytes()
	_ = NewTextPreprocessor(str).DO()
	// fmt.Println(preprocessor.Lemmas)
}

func TestMultiThreadPreprocessing(t *testing.T) {
	str:= ReadBytes()
	_ = NewTextPreprocessor(str).DOPullThread(5)
	// fmt.Println(preprocessor.Lemmas)
}

func BenchmarkPullThreadPreprocessing(b *testing.B) {
	b.ReportAllocs()
	str:= ReadBytes()
	preprocessor := NewTextPreprocessor(str)
	//
	for i := 0; i < b.N; i++ {
		preprocessor.DOPullThread( 5)
	}
}

func BenchmarkOneThreadPreprocessing(b *testing.B) {
	b.ReportAllocs()
	str:= ReadBytes()
	preprocessor := NewTextPreprocessor(str)
	//
	for i := 0; i < b.N; i++ {
		preprocessor.DO()
	}
}
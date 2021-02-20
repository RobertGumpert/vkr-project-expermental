package text_preprocessing

import (
	"encoding/json"
	"fmt"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/bbalet/stopwords"
	concurrent_map "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/runtimeinfo"
	"math"
	"regexp"
	"strings"
	"sync"
)

type TextPreprocessor struct {
	rwMutex    *sync.RWMutex
	wg         *sync.WaitGroup
	OriginText *string `json:"-"`
	//
	ClearText *string   `json:"-"`
	Words     *[]string `json:"words"`
	Lemmas    *[]string `json:"lemmas"`
	Stems     *[]string `json:"stems"`
	//
	// Частота лемм и стемм
	//
	LemmasFrequency *concurrent_map.ConcurrentMap `json:"lemmas_frequency"`
	StemsFrequency  *concurrent_map.ConcurrentMap `json:"stems_frequency"`
	//
	// Частота биграмм составленных из лемм и стемм
	//
	LemmasBigramFrequency *concurrent_map.ConcurrentMap `json:"lemmas_bigram_frequency"`
	StemsBigramFrequency  *concurrent_map.ConcurrentMap `json:"stems_bigram_frequency"`
	//
	// Mutual Information.
	// Коэффициент посчитанный для каждой биграммы, показывает
	// статистическую значимость каждого слова биграммы в сравнении
	// с их совместной частотой. То есть:
	// 		* Если 1 < MI, то биграмма значима.
	// 		* Если 0 < MI < 1, то биграмма не значима.
	// 		* Если MI < 0, то каждое слово, может встречаться только тогда, когда другое никогда не встречается.
	//
	LemmasMI *concurrent_map.ConcurrentMap `json:"lemmas_mi"`
	StemsMI  *concurrent_map.ConcurrentMap `json:"stems_mi"`
}

func NewTextPreprocessor(str string) *TextPreprocessor {
	str = strings.ToLower(str)
	//
	regexAscii := regexp.MustCompile("[[:^ascii:]]")
	str = regexAscii.ReplaceAllLiteralString(str, " ")
	//
	regexCodeFunctions := regexp.MustCompile(`(?i)[\w\d]+[.](?i)[\w\d]+[(](?i)[\w\d]{0,}[)]`)
	str = regexCodeFunctions.ReplaceAllString(str, " ")
	//
	regexSymbols, _ := regexp.Compile(`[]\d%:$"';[&*=<>}{)(?!/.,\-]`)
	str = regexSymbols.ReplaceAllString(str, " ")
	//
	var (
		preprocessor = new(TextPreprocessor)
		clearText    = stopwords.CleanString(str, "en", true)
		words        = make([]string, 0)
	)
	words = strings.Fields(clearText)
	//
	preprocessor.Words = &words
	preprocessor.ClearText = &clearText
	//
	lemmasSlice, stemsSlice := make([]string, len(words)), make([]string, len(words))
	preprocessor.Lemmas, preprocessor.Stems = &lemmasSlice, &stemsSlice
	//
	lemmasFrequency, stemsFrequency,
	lemmasBigramFrequency, stemsBigramFrequency,
	lemmasMi, stemsMi := concurrent_map.New(), concurrent_map.New(),
		concurrent_map.New(), concurrent_map.New(),
		concurrent_map.New(), concurrent_map.New()
	//
	preprocessor.LemmasFrequency, preprocessor.StemsFrequency,
		preprocessor.LemmasBigramFrequency, preprocessor.StemsBigramFrequency,
		preprocessor.LemmasMI, preprocessor.StemsMI = &lemmasFrequency, &stemsFrequency,
		&lemmasBigramFrequency, &stemsBigramFrequency,
		&lemmasMi, &stemsMi
	//
	return preprocessor
}

func (preprocessor *TextPreprocessor) DO() *TextPreprocessor {
	preprocessor.do(
		0,
		int64(len(*preprocessor.Words)),
		false,
	)
	preprocessor.mutualInformation()
	return preprocessor
}

func (preprocessor *TextPreprocessor) DOPullThread(pullThreadSize int) *TextPreprocessor {
	var (
		lengthPieceOfArray = len(*preprocessor.Words) / pullThreadSize
		do                 = func(from, to int64, preprocessor *TextPreprocessor) {
			preprocessor.wg.Add(1)
			go preprocessor.do(
				from,
				to,
				true,
			)
		}
	)
	//
	preprocessor.rwMutex = new(sync.RWMutex)
	preprocessor.wg = new(sync.WaitGroup)
	//
	if lengthPieceOfArray < 1 || pullThreadSize <= 1 {
		do(
			0,
			int64(len(*preprocessor.Words)),
			preprocessor,
		)
	} else {
		var (
			to   = lengthPieceOfArray
			from = 0
		)
		for ; (from + lengthPieceOfArray) < len(*preprocessor.Words); from, to = from+lengthPieceOfArray, to+lengthPieceOfArray {
			do(
				int64(from),
				int64(to),
				preprocessor,
			)
		}
		if from < len(*preprocessor.Words) {
			do(
				int64(from),
				int64(len(*preprocessor.Words)),
				preprocessor,
			)
		}
	}
	preprocessor.wg.Wait()
	preprocessor.mutualInformation()
	return preprocessor
}

func (preprocessor *TextPreprocessor) do(from, to int64, waitGroupDone bool) {
	if waitGroupDone {
		defer preprocessor.wg.Done()
	}
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), err)
		return
	}
	for ; from < to; from++ {

		//if len((*preprocessor.Words)[from]) == 1 {
		//	(*preprocessor.Words)[from] = (*preprocessor.Words)[len(*preprocessor.Words)-1]
		//	(*preprocessor.Words)[len(*preprocessor.Words)-1] = ""
		//	*preprocessor.Words = (*preprocessor.Words)[:len(*preprocessor.Words)-1]
		//	to--
		//	continue
		//}

		var (
			currentLemma = GetLemma(&((*preprocessor.Words)[from]), lemmatizer)
			currentStem  = GetStem(&((*preprocessor.Words)[from]))
		)
		//
		// Write the word stem and the word lemma
		//
		(*preprocessor.Stems)[from] = *currentStem
		(*preprocessor.Lemmas)[from] = *currentLemma
		//
		// Create a bigram from the current and next lemma or stem.
		//
		var (
			lemmasBigram string
			stemBigram   string
			nextLemma    *string
			nextStem     *string
		)
		if from <= to-1 && (from+1) != int64(len(*preprocessor.Words)) {
			nextLemma = GetLemma(&((*preprocessor.Words)[from+1]), lemmatizer)
			nextStem = GetStem(&((*preprocessor.Words)[from+1]))
			lemmasBigram = strings.Join([]string{
				*currentLemma,
				*nextLemma,
			}, " ")
			stemBigram = strings.Join([]string{
				*currentStem,
				*nextStem,
			}, " ")
		}
		//
		// Calculate frequencies
		//
		writeToFrequencyMap(currentLemma, preprocessor.LemmasFrequency)
		writeToFrequencyMap(currentStem, preprocessor.StemsFrequency)
		writeToFrequencyMap(&lemmasBigram, preprocessor.LemmasBigramFrequency)
		writeToFrequencyMap(&stemBigram, preprocessor.StemsBigramFrequency)
	}
	return
}

func (preprocessor *TextPreprocessor) mutualInformation() {
	var (
		localWaitGroup = new(sync.WaitGroup)
		calcMi         = func(xy, x, y int64, n int) float64 {
			num := (float64(xy) * float64(n)) / (float64(x) * float64(y))
			return math.Log2(num)
		}
		doMI = func(bigramFrequency, wordFrequency, miFrequency *concurrent_map.ConcurrentMap, n int, localWaitGroup *sync.WaitGroup) {
			defer localWaitGroup.Done()
			for item := range bigramFrequency.IterBuffered() {
				if strings.TrimSpace(item.Key) == "" {
					continue
				}
				xy := item.Val
				splitBigram := strings.Split(item.Key, " ")
				x, _ := wordFrequency.Get(splitBigram[0])
				y, _ := wordFrequency.Get(splitBigram[1])
				mi := calcMi(
					xy.(int64),
					x.(int64),
					y.(int64),
					n,
				)
				miFrequency.Set(item.Key, mi)
			}
		}
		n = len(*preprocessor.Words)
	)
	if preprocessor.LemmasBigramFrequency.Count() != 0 {
		localWaitGroup.Add(1)
		go doMI(
			preprocessor.LemmasBigramFrequency,
			preprocessor.LemmasFrequency,
			preprocessor.LemmasMI,
			n,
			localWaitGroup,
		)
	}
	if preprocessor.StemsBigramFrequency.Count() != 0 {
		localWaitGroup.Add(1)
		go doMI(
			preprocessor.StemsBigramFrequency,
			preprocessor.StemsFrequency,
			preprocessor.StemsMI,
			n,
			localWaitGroup,
		)
	}
	localWaitGroup.Wait()
	return
}

func (preprocessor *TextPreprocessor) ToString() string {
	return strings.Join([]string{
		fmt.Sprintf("Clear text is : %v", *preprocessor.Words),
		"---",
		fmt.Sprintf("Lemmas is : %v", *preprocessor.Lemmas),
		"---",
		fmt.Sprintf("Stems is : %v", *preprocessor.Stems),
		"---",
		fmt.Sprintf("Lemmas Frequency is : %v", *preprocessor.LemmasFrequency),
		"---",
		fmt.Sprintf("Stems Frequency is : %v", *preprocessor.StemsFrequency),
		"---",
		fmt.Sprintf("MI Lemmas is : %v", *preprocessor.LemmasBigramFrequency),
	}, "\n")
}

func (preprocessor *TextPreprocessor) Serialize() ([]byte, string, error) {
	bs, err := json.Marshal(preprocessor)
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}
	return bs, string(bs), nil
}

func writeToFrequencyMap(key *string, mp *concurrent_map.ConcurrentMap) {
	if val, ok := mp.Get(*key); ok {
		v := val.(int64)
		v++
		mp.Set(*key, v)
	} else {
		mp.Set(*key, int64(1))
	}
}

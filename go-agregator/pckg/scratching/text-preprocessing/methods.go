package text_preprocessing

import (
	"fmt"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/kljensen/snowball"
	concurrentMap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/runtimeinfo"
	"math"
	"sync"
)

func EuclideanDistance(vecA, vecB *[]float64) (D float64) {
	var (
		sum = float64(0)
	)
	for i := 0; i < len(*vecA); i++ {
		pow := math.Pow((*vecA)[i]-(*vecB)[i], float64(2))
		sum += pow
	}
	return math.Sqrt(sum)
}

func MethodSetOperations(crossingA, crossingB, divergenceA, divergenceB *concurrentMap.ConcurrentMap) (Kab, KnotAB float64) {
	var (
		//s          = float64(0)
		ka, knota                              = float64(0), float64(0)
		kb, knotb                              = float64(0), float64(0)
		crossLenA, crossLenB, divLenA, divLenB = crossingA.Count(), crossingB.Count(), divergenceA.Count(), divergenceB.Count()
		lenA, lenB                             = float64(crossLenA + divLenA), float64(crossLenB + divLenB)
		sumFrequency                           = func(vec *concurrentMap.ConcurrentMap) float64 {
			if vec.Count() == 0 {
				return 0
			}
			sum := float64(0)
			for item := range vec.IterBuffered() {
				switch item.Val.(type) {
				case int64:
					sum += float64(item.Val.(int64))
				case float64:
					sum += item.Val.(float64)
				}
			}
			return sum // / float64(vec.Count())
		}
	)
	if crossLenA == 0 && crossLenB == 0 &&
		divLenA == 0 && divLenB == 0 {
		return 0, 0
	}
	if crossLenA == 0 || crossLenB == 0 {
		return 0, 0
	}
	//if (divLenA == divLenB) &&
	//	(divLenA == 0 && crossLenA != 0) {
	//	return 1, 0
	//}
	//
	ka = float64(crossLenA) / lenA
	kb = float64(crossLenB) / lenB
	knota = (lenA - float64(crossLenA)) / lenA
	knotb = (lenB - float64(crossLenB)) / lenB
	//
	var (
		fine = float64(0)
	)
	if lenA != lenB {
		fine = (math.Abs(lenB - lenA)) / (lenA + lenB)
	}
	//
	var (
		kab    = (ka + kb) / (2 + fine)
		knotab = (knota + knotb) / (2 + fine)
	)
	wa := []float64{
		sumFrequency(crossingA),
		sumFrequency(divergenceA),
	}
	wb := []float64{
		sumFrequency(crossingB),
		sumFrequency(divergenceB),
	}
	var printMap = func(key string, vec *concurrentMap.ConcurrentMap) string {
		var str = fmt.Sprintf("%s = [ ", key)
		for item := range vec.IterBuffered() {
			str += fmt.Sprintf("%s [%d], ", item.Key, item.Val)
		}
		return str + "] "
	}
	fmt.Println(
		fmt.Sprintf("\t\t%s\n", printMap("cross A", crossingA)),
		fmt.Sprintf("\t\t%s\n", printMap("cross B", crossingB)),
		fmt.Sprintf("\t\t%s\n", printMap("div   A", divergenceA)),
		fmt.Sprintf("\t\t%s", printMap("div   B", divergenceB)),
	)
	fmt.Println(
		fmt.Sprintf("\t\t"+
			"Cross len. = %d, "+
			"\tK(a) = %0.02f,"+
			"\tK(b) = %0.02f,"+
			"\t!K(a) = %0.02f,"+
			"\t!K(b) = %0.02f,"+
			"\tK(ab) = %.02f,"+
			"\t!K(ab) = %.02f"+
			"\tWa = {%0.02f ; %0.02f}"+
			"\tWb = {%0.02f ; %0.02f}",
			crossLenA,
			ka,
			kb,
			knota,
			knotb,
			kab,
			knotab,
			wa[0], wa[1],
			wb[0], wb[1],
		),
	)
	return kab, knotab
}

func GetCrossing(vecA, vecB *concurrentMap.ConcurrentMap) (AB, NotAB int64, CrossingA, CrossingB, DivergenceA, DivergenceB *concurrentMap.ConcurrentMap) {
	var (
		crossingA, crossingB     = concurrentMap.New(), concurrentMap.New()
		divergenceA, divergenceB = concurrentMap.New(), concurrentMap.New()
	)
	for item := range vecA.IterBuffered() {
		if val, exist := vecB.Get(item.Key); exist {
			AB++
			crossingA.Set(item.Key, item.Val)
			crossingB.Set(item.Key, val)
		} else {
			divergenceA.Set(item.Key, item.Val)
		}
	}
	for item := range vecB.IterBuffered() {
		if !crossingB.Has(item.Key) {
			divergenceB.Set(item.Key, item.Val)
		}
	}
	NotAB = (int64(vecA.Count()) + int64(vecB.Count())) - AB
	return AB, NotAB, &crossingA, &crossingB, &divergenceA, &divergenceB
}

func CosineDistance(vecA, vecB *concurrentMap.ConcurrentMap) float64 {
	var (
		gwg           = new(sync.WaitGroup)
		calcNumerator = func(vecAC, vecBC *[]float64, result *float64, gwg *sync.WaitGroup) {
			defer gwg.Done()
			var s float64
			for i := 0; i < len(*vecAC); i++ {
				s += (*vecAC)[i] * (*vecBC)[i]
			}
			*result = s
		}
		calcDenominator = func(vecAC, vecBC *[]float64, result *float64, gwg *sync.WaitGroup) {
			defer gwg.Done()
			var (
				sA          float64
				sB          float64
				calcSummary = func(vec *[]float64, result *float64, wg *sync.WaitGroup) {
					defer wg.Done()
					var s float64
					for i := 0; i < len(*vec); i++ {
						s += (*vec)[i] * (*vec)[i]
					}
					*result = s
				}
				wg = new(sync.WaitGroup)
			)
			wg.Add(2)
			go calcSummary(vecAC, &sA, wg)
			go calcSummary(vecBC, &sB, wg)
			wg.Wait()
			*result = math.Sqrt(sA) * math.Sqrt(sB)
		}
	)
	vecAC, vecBC := Vectorized(vecA, vecB)
	gwg.Add(2)
	var (
		n, d float64
	)
	go calcNumerator(vecAC, vecBC, &n, gwg)
	go calcDenominator(vecAC, vecBC, &d, gwg)
	gwg.Wait()
	return n / d
}

func GetStem(word *string) (stem *string) {
	stemm, err := snowball.Stem(*word, "english", true)
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), err)
		return word
	}
	return &stemm
}

func GetLemma(word *string, lemmatizers ...*golem.Lemmatizer) (lemma *string) {
	var lemmatizer *golem.Lemmatizer
	if lemmatizers == nil || len(lemmatizers) == 0 {
		lem, err := golem.New(en.New())

		if err != nil {
			fmt.Println(runtimeinfo.Runtime(1), err)
			return word
		}
		lemmatizer = lem
	} else {
		lemmatizer = lemmatizers[0]
	}
	result := lemmatizer.Lemma(*word)
	return &result
}

func LemmingProcessor(words, result *[]string, calcFrequency bool) (frequency *map[string]int64) {
	if len(*words) != len(*result) {
		fmt.Println(runtimeinfo.Runtime(1), " len's slices not equals.")
		return
	}
	if calcFrequency {
		f := make(map[string]int64, 0)
		frequency = &f
	}
	lem, err := golem.New(en.New())
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), err)
		return
	}
	for i := 0; i < len(*words); i++ {
		(*result)[i] = lem.Lemma((*words)[i])
		if calcFrequency {
			if _, isExist := (*frequency)[(*result)[i]]; isExist {
				(*frequency)[(*result)[i]]++
			} else {
				(*frequency)[(*result)[i]] = 1
			}
		}
	}
	return
}

func StemsProcessor(words, result *[]string, calcFrequency bool) (frequency *map[string]int64) {
	if len(*words) != len(*result) {
		fmt.Println(runtimeinfo.Runtime(1), " len's slices not equals.")
		return
	}
	if calcFrequency {
		f := make(map[string]int64, 0)
		frequency = &f
	}
	for i := 0; i < len(*words); i++ {
		stem, err := snowball.Stem((*words)[i], "english", true)
		if err != nil {
			(*result)[i] = (*words)[i]
		} else {
			(*result)[i] = stem
		}
		if calcFrequency {
			if _, isExist := (*frequency)[(*result)[i]]; isExist {
				(*frequency)[(*result)[i]]++
			} else {
				(*frequency)[(*result)[i]] = 1
			}
		}
	}
	return
}

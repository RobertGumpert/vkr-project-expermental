package main

import (
	"bufio"
	"fmt"
	cmap "github.com/streamrail/concurrent-map"
	text_preprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	dictFile  = "repositories-dictionary.txt"
	repoFiles = []string{
		"react.txt",
		"angular.txt",
		"vue.txt",
		"gin.txt",
		"flask.txt",
		"okhttp.txt",
		"hyper.txt",
		"terminal.txt",
		"alacritty.txt",
	}
	text1 = "When all else is quiet When quiet quiet"
	text2 = "When is she supposed to bring When supposed to"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	runtime.GOMAXPROCS(runtime.NumCPU())

	//_, terminalTitles, _ := getFileContent("terminal.txt")
	//_, alacrittyTitles, _ := getFileContent("alacritty.txt")

	//dict := concurrentMap.New()
	//dict.Set("quiet", int64(0))

	writeDictionaryAndVectorizedCorpus(dictFile, []string{
		// "react.txt",
		//"angular.txt",
		//"vue.txt",
		"gin.txt",
		// "flask.txt",
		//"okhttp.txt",
		// "hyper.txt",
		//"terminal.txt",
		//"alacritty.txt",
	})

	//createAndWriteDictionary(dictFile, []string{
	//	// "react.txt",
	//	//"angular.txt",
	//	"vue.txt",
	//	"gin.txt",
	//	// "flask.txt",
	//	"okhttp.txt",
	//	// "hyper.txt",
	//	"terminal.txt",
	//	"alacritty.txt",
	//})

	//p1 := text_preprocessing.NewTextPreprocessor(text1).DO()
	//p2 := text_preprocessing.NewTextPreprocessor(text2).DO()
	//////
	//model1 := &text_preprocessing.VectorizedCorpusModel{
	//	Key:            "terminal",
	//	FrequencyWords: p1.LemmasFrequency,
	//}
	//model2 := &text_preprocessing.VectorizedCorpusModel{
	//	Key:            "alacritty",
	//	FrequencyWords: p2.LemmasFrequency,
	//}
	////
	//slice, mp := text_preprocessing.CreateDictionaryFromCorpus(
	//	model1,
	//	model2,
	//)
	////
	//vectorModel1 := text_preprocessing.VectorizedCorpus(
	//	model1,
	//	model2,
	//)
	////
	//vectorModel2 := text_preprocessing.VectorizedWithDictionary(
	//	&dict,
	//	model1,
	//	model2,
	//)
	//
	//fmt.Println(vectorModel1)
	// fmt.Println(vectorModel2)
	//fmt.Println(slice)
	//fmt.Println(mp)
	//
	//getFileContent("terminal.txt")
	fmt.Println("Final...")
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if scanner.Text() == "end" {
			os.Exit(1)
		}
	}
	return
}

func createAndWriteDictionary(fileDict, fileVec string, repos []string) {
	var (
		mx     = new(sync.Mutex)
		wg     = new(sync.WaitGroup)
		corpus = make([]*text_preprocessing.VectorizedCorpusModel, 0)
	)
	for _, repo := range repos {
		wg.Add(1)
		go func(repo string, corpus *[]*text_preprocessing.VectorizedCorpusModel, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			fileContent := getFileContent(repo, true)
			mx.Lock()
			*corpus = append(*corpus, &text_preprocessing.VectorizedCorpusModel{
				Key:            repo,
				FrequencyWords: fileContent.TitlesPreprocessing.LemmasFrequency,
			})
			mx.Unlock()
			return
		}(repo, &corpus, wg, mx)

	}
	wg.Wait()
	sliceDict, _ := text_preprocessing.CreateDictionaryFromCorpus(corpus...)
	writeContent := make([]string, 0)
	for i := 0; i < len(*sliceDict); i++ {
		if *(*sliceDict)[i] == "" {
			fmt.Println("empty.")
			continue
		}
		writeContent = append(
			writeContent,
			strings.Join(
				[]string{
					*(*sliceDict)[i],
					strconv.Itoa(i),
				},
				" *-* ",
			),
		)
	}
	path := getDirPath(fileDict)
	output := []byte(strings.Join(writeContent, "\n"))
	err := ioutil.WriteFile(path, output, 0644)
	if err != nil {
		panic(err)
	}
}

func writeDictionaryAndVectorizedCorpus(fileDict string, repos []string) {
	path := getDirPath(fileDict)
	bts, err := ioutil.ReadFile(path)
	//
	if err != nil {
		panic(err)
	}
	//
	var (
		mx         = new(sync.Mutex)
		wg         = new(sync.WaitGroup)
		corpus     = make([]*text_preprocessing.VectorizedCorpusModel, 0)
		dictionary = cmap.New()
	)
	wg.Add(1)
	go func(dictionary *cmap.ConcurrentMap, bts []byte, wg *sync.WaitGroup) {
		defer wg.Done()
		content := strings.Split(string(bts), "\n")
		for i := 0; i < len(content); i++ {
			word := strings.Split(content[i], " *-* ")
			key := word[0]
			value, err := strconv.ParseInt(word[1], 10, 64)
			if err != nil {
				panic(err)
			}
			dictionary.Set(key, value)
		}
		return
	}(&dictionary, bts, wg)
	for _, repo := range repos {
		wg.Add(1)
		go func(repo string, corpus *[]*text_preprocessing.VectorizedCorpusModel, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			fileContent := getFileContent(repo, false)
			c := make([]*text_preprocessing.VectorizedCorpusModel, 0)
			sep := ""
			if strings.Contains(*fileContent.Titles, "\r\n") {
				sep = "\r\n"
			}
			if strings.Contains(*fileContent.Titles, "\n") {
				sep = "\n"
			}
			repoName := strings.Split(repo, ".txt")[0]
			titles := strings.Split(*fileContent.Titles, sep)
			for i, title := range titles {
				titlePreprocessor := text_preprocessing.NewTextPreprocessor(title).DO()
				key := repoName + "_issue_" + strconv.Itoa(i)
				c = append(c, &text_preprocessing.VectorizedCorpusModel{
					Key:            key,
					FrequencyWords: titlePreprocessor.LemmasFrequency,
				})
				fmt.Println("Repo: ", repo, ". issue: ", i)
			}
			fmt.Println("Repo: ", repo, ". copy...")
			mx.Lock()
			*corpus = append(*corpus, c...)
			mx.Unlock()
			fmt.Println("Repo: ", repo, ". finish.")
			return
		}(repo, &corpus, wg, mx)

	}
	wg.Wait()
	//
	model := text_preprocessing.VectorizedWithDictionary(
		&dictionary,
		corpus...,
	)
	fmt.Println(model)
	//
	//for item := range model.GetPresenceVectors().IterBuffered() {
	//	val := item.Val.()
	//}
}

package main

import (
	"bufio"
	"fmt"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	cmap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/scratching/textProcessor/textClearing"
	"go-agregator/pckg/scratching/textProcessor/textMetrics"
	"go-agregator/pckg/scratching/textProcessor/textVectoring"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	learningVectorsFile = "learning-vectors.csv"
	dictionaryFile      = "repositories-dictionary.txt"
	repositoriesFiles   = []string{
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
	e := en.New()
	lemmatizer, _ := golem.New(e)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic: ", r)
		}
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			if scanner.Text() == "end" {
				os.Exit(1)
			}
		}
	}()

	//str := "Hello, мир сделай f myClass.myMethod(a), MyClass.MyMethod(a) / myclass.mymethod(a) fast? [Yes] or [no]? d f g"
	//str := "Hello f / \r\n? [Yes] or [no]? d f g Hello-my-friend f / \r\n? [Yes] or [No]? d f g [yes] or [no]"
	str := "refactor(browser): merge static & dynamic platforms"

	doClearTopics := textClearing.CustomClear(
		false,
		lemmatizer,
		[]textClearing.Contains{
			textClearing.ContainsCode,
		},
		[]textClearing.Clear{
			textClearing.ClearASCII,
			textClearing.ClearSymbols,
		},
	)
	doClearTitles := textClearing.CustomClear(
		false,
		lemmatizer,
		[]textClearing.Contains{
			textClearing.ContainsCode,
		},
		[]textClearing.Clear{
			textClearing.ClearSpecialWord,
			textClearing.ClearASCII,
			textClearing.ClearSymbols,
		},
	)
	doClearBodies := textClearing.CustomClear(
		false,
		lemmatizer,
		[]textClearing.Contains{
			textClearing.ContainsCode,
		},
		[]textClearing.Clear{
			textClearing.ClearSpecialWord,
			textClearing.ClearASCII,
			textClearing.ClearSymbols,
		},
	)

	_, _, _ = doClearTitles(&str)
	_, _, _ = doClearTopics(&str)
	_, _, _ = doClearBodies(&str)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(*newStr)
	//fmt.Println(*slice)
	//cmp := textMetrics.WordsFrequency(slice)
	//for item := range cmp.IterBuffered() {
	//	fmt.Println(item.Key, " = ", item.Val)
	//}

	//
	//lemming(repositoriesFiles, "topics", lemmatizer, doClearTopics)
	lemming(repositoriesFiles, "titles", lemmatizer, doClearTitles)
	//lemming(repositoriesFiles, "descriptions", lemmatizer, doClearTopics)
	//lemming(repositoriesFiles, "bodies", lemmatizer, doClearBodies)
	//
	//tfidf(repositoriesFiles, "topics")
	tfidf(repositoriesFiles, "titles")
	//tfidf(repositoriesFiles, "descriptions")
	//tfidf(repositoriesFiles, "bodies")
	//
	//r := "result del_if [non lem, code] clear [ascii, symbols, spec. words]"
	//result := "result del_if [code] clear [ascii, symbols, spec. words]"
	cosineDistance(repositoriesFiles, "titles", "result")
	//cosineDistance(repositoriesFiles, "topics", result)
	//cosineDistance(repositoriesFiles, "descriptions", result)
	//cosineDistance(repositoriesFiles, "descriptions", "result")

	fmt.Println("Final...")
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if scanner.Text() == "end" {
			os.Exit(1)
		}
	}
	return
}

func tfidf(repositoriesFiles []string, group string) {
	root := "group-by-metrics/tf-idf/" + group
	var getFile = func(repo string, group string) (string, string) {
		name := strings.Split(repo, ".txt")[0]
		f := name + "-" + group + ".txt"
		groupingFile := getDirPath(f, "group-by-clear/clear-by-stop-words-lemmas/"+group)
		content, _ := ReadBytes(groupingFile)
		return content, f
	}
	//
	mp := make(map[int]string, 0)
	documents := make([]*[]string, len(repositoriesFiles))
	fmt.Println("Get documents...")
	for index, repo := range repositoriesFiles {
		content, fileName := getFile(repo, group)
		mp[index] = fileName
		if len(content) == 0 {
			continue
		}
		sep := ""
		if strings.Contains(content, "\r\n") {
			sep = "\r\n"
		}
		if strings.Contains(content, "\n") {
			sep = "\n"
		}
		documentContent := make([]string, 0)
		if sep == "" {
			slice := textClearing.ToSlice(&content)
			documentContent = append(documentContent, *slice...)
		} else {
			elements := strings.Split(content, sep)
			for _, element := range elements {
				slice := textClearing.ToSlice(&element)
				if len(*slice) == 0 {
					continue
				}
				documentContent = append(documentContent, *slice...)
			}
		}
		documents[index] = &documentContent
	}
	fmt.Println("Documents ready!")
	wordsIDF, documentsTF, dictionary := textMetrics.GetTFIDFMetrics(&documents)
	//
	fmt.Println("Metrics ready!")
	writeContentDict := make([]string, 0)
	writeContentIDF := make([]string, 0)
	for item := range dictionary.IterBuffered() {
		strDict := item.Key + " *=* " + strconv.Itoa(int(item.Val.(int64)))
		writeContentDict = append(writeContentDict, strDict)
		//
		val, _ := wordsIDF.Get(item.Key)
		strIDF := item.Key + " *=* " + fmt.Sprintf("%f", val.(float64))
		writeContentIDF = append(writeContentIDF, strIDF)
	}
	path := getDirPath("dict.txt", root)
	output := []byte(strings.Join(writeContentDict, "\n"))
	err := ioutil.WriteFile(path, output, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Write dict!")
	path = getDirPath("idf.txt", root)
	output = []byte(strings.Join(writeContentIDF, "\n"))
	err = ioutil.WriteFile(path, output, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Write idf!")
	//
	for i := 0; i < len(*documentsTF); i++ {
		tfidf, err := textMetrics.TFIDF((*documentsTF)[i], wordsIDF)
		fileName := mp[i]
		path := getDirPath(fileName, root)
		if err != nil {
			output = []byte("")
			err = ioutil.WriteFile(path, output, 0644)
			if err != nil {
				panic(err)
			}
			fmt.Println("Write NIL ", fileName, " !")
			continue
		}
		writeContent := make([]string, 0)
		for item := range tfidf.IterBuffered() {
			str := item.Key + " *=* " + fmt.Sprintf("%f", item.Val.(float64))
			writeContent = append(writeContent, str)
		}
		output = []byte(strings.Join(writeContent, "\n"))
		err = ioutil.WriteFile(path, output, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("Write ", fileName, " !")
	}
	fmt.Println("OK!")
}

func cosineDistance(repositoriesFiles []string, group, file string) {
	rootWrite := "group-by-metrics/cosine/tf-idf/" + group
	rootRead := "group-by-metrics/tf-idf/" + group
	//
	path := getDirPath("dict.txt", rootRead)
	content, _ := ReadBytes(path)
	documentContent := strings.Split(content, "\n")
	dictionary := cmap.New()
	for _, row := range documentContent {
		s := strings.Split(row, " *=* ")
		s[0] = strings.TrimSpace(s[0])
		s[1] = strings.TrimSpace(s[1])
		if s[0] == "" || s[1] == "" {
			fmt.Println("Error: empty string s[0] = '", s[0], "' or s[1] = '", s[1], "' ")
			continue
		}
		//
		if strings.Contains(s[1], "\r") {
			s[1] = strings.Split(s[1], "\r")[0]
		}
		index, err := strconv.Atoi(s[1])
		if err != nil {
			fmt.Println(err)
			continue
		}
		dictionary.Set(s[0], index)
	}
	//
	var getFile = func(repo string, group string) (string, string) {
		name := strings.Split(repo, ".txt")[0]
		f := name + "-" + group + ".txt"
		groupingFile := getDirPath(f, rootRead)
		content, _ := ReadBytes(groupingFile)
		return content, f
	}
	//
	mp := make(map[int]string)
	vectors := make([]*cmap.ConcurrentMap, len(repositoriesFiles))
	for i, repo := range repositoriesFiles {
		content, _ := getFile(repo, group)
		documentContent := strings.Split(content, "\n")
		tfidfDocument := cmap.New()
		for _, row := range documentContent {
			s := strings.Split(row, " *=* ")
			s[0] = strings.TrimSpace(s[0])
			s[1] = strings.TrimSpace(s[1])
			if s[0] == "" || s[1] == "" {
				fmt.Println("Error: empty string s[0] = '", s[0], "' or s[1] = '", s[1], "' ")
				continue
			}
			//
			if strings.Contains(s[1], "\r") {
				s[1] = strings.Split(s[1], "\r")[0]
			}
			index, err := strconv.ParseFloat(s[1], 64)
			if err != nil {
				fmt.Println(err)
				continue
			}
			tfidfDocument.Set(s[0], index)
		}
		vectors[i] = &tfidfDocument
		mp[i] = repo
	}
	err := textVectoring.Vectorized(&dictionary, &vectors)
	if err != nil {
		panic(err)
	}
	//
	writeContent := make([]string, 0)
	for i := range repositoriesFiles {
		str := "\nMAIN: " + mp[i]
		for j := range repositoriesFiles {
			str += "\n\t" + mp[i] + " -> " + mp[j] + " : "
			dist, err := textMetrics.CosineDistance(vectors[i], vectors[j])
			if err != nil {
				str += err.Error()
			}
			str += fmt.Sprintf("%.002f", dist)
		}
		fmt.Println(str)
		writeContent = append(writeContent, str)
	}
	path = getDirPath(file+".txt", rootWrite)
	output := []byte(strings.Join(writeContent, "\n"))
	err = ioutil.WriteFile(path, output, 0644)
	if err != nil {
		panic(err)
	}
	return
}

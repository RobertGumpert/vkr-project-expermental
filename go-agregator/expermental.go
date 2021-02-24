package main

import (
	"encoding/csv"
	"fmt"
	"github.com/aaaton/golem/v4"
	cmap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/runtimeinfo"
	issuesCollector "go-agregator/pckg/scratching/github-api/issues-collector"
	text_preprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"go-agregator/pckg/scratching/textProcessor/textClearing"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//
// "./test.txt"
//
func ReadBytes(path string) (string, []byte) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return "", nil
	}
	return string(content), content
}

func deserialize(file string) issuesCollector.List {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
	if err != nil {
		fmt.Println(runtimeinfo.Runtime(1), ", ERROR: ", err)
		return nil
	}
	_, content := ReadBytes(absPath)
	return issuesCollector.DeserializeToList(content)
}

type FileContent struct {
	DescriptionPreprocessing,
	TitlesPreprocessing,
	BodiesPreprocessing *text_preprocessing.TextPreprocessor
	//
	Description,
	Topics,
	Titles,
	Bodies *string
}

func getFileContent(file string, doPreprocessing bool) *FileContent {
	var (
		wg = new(sync.WaitGroup)
		content                             = readFile(file)
		description, topics, titles, bodies = separateContent(content)
		descriptionTopicsContent            = strings.Join([]string{*description, *topics}, " ")
		descriptionPreprocessing *text_preprocessing.TextPreprocessor
		titlesPreprocessing      *text_preprocessing.TextPreprocessor
		bodiesPreprocessing      *text_preprocessing.TextPreprocessor
		do = func(content *string, preprocessor *text_preprocessing.TextPreprocessor, wg *sync.WaitGroup) {
			defer wg.Done()
			*preprocessor = *text_preprocessing.NewTextPreprocessor(*content).DOPullThread(5)
			return
		}
	)
	if doPreprocessing {
		wg.Add(2)
		descriptionPreprocessing = new(text_preprocessing.TextPreprocessor)
		titlesPreprocessing = new(text_preprocessing.TextPreprocessor)
		//bodiesPreprocessing      = new(text_preprocessing.TextPreprocessor)
		go do(&descriptionTopicsContent, descriptionPreprocessing, wg)
		go do(titles, titlesPreprocessing, wg)
		wg.Wait()
	}
	return &FileContent{
		DescriptionPreprocessing: descriptionPreprocessing,
		TitlesPreprocessing:      titlesPreprocessing,
		BodiesPreprocessing:      bodiesPreprocessing,
		Description:              description,
		Topics:                   topics,
		Titles:                   titles,
		Bodies:                   bodies,
	}
}

func separateContent(content string) (*string, *string, *string, *string) {
	var (
		wg = new(sync.WaitGroup)
		//
		titlesSeparator      = getTextSeparator(content, "Titles")
		bodiesSeparator      = getTextSeparator(content, "Bodies")
		descriptionSeparator = getTextSeparator(content, "Description")
		topicsSeparator      = getTextSeparator(content, "Topics")
		//
		titles, bodies, description, topics = "", "", "", ""
	)
	wg.Add(4)
	//
	go strippedWithTwoSeparator(&description, &descriptionSeparator, &topicsSeparator, content, wg)
	go strippedWithTwoSeparator(&topics, &topicsSeparator, &titlesSeparator, content, wg)
	go strippedWithTwoSeparator(&titles, &titlesSeparator, &bodiesSeparator, content, wg)
	go strippedWithOneSeparator(&bodies, &bodiesSeparator, content, wg)
	//
	wg.Wait()
	return &description, &topics, &titles, &bodies
}

func strippedWithTwoSeparator(output, firstSeparator, secondSeparator *string, content string, wg *sync.WaitGroup) {
	defer wg.Done()
	firstSlice := strings.Split(content, *firstSeparator)
	secondSlice := strings.Split(
		firstSlice[1],
		*secondSeparator,
	)
	*output = secondSlice[0]
}

func strippedWithOneSeparator(bodies, bodiesSeparator *string, content string, wg *sync.WaitGroup) {
	defer wg.Done()
	slice := strings.Split(
		content,
		*bodiesSeparator,
	)
	*bodies = slice[1]
}

func getDirPath(file, dir string) string {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/%s/%s", dir, file))
	if err != nil {
		applicationCrashed(err)
		return file
	}
	return absPath
}

func readFile(file string) string {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
	if err != nil {
		applicationCrashed(err)
		return file
	}
	content, _ := ReadBytes(absPath)
	return content
}

func getTextSeparator(content, separator string) string {
	var (
		windowsSeparator = osTextTemplate(separator, "w")
		linuxSeparator   = osTextTemplate(separator, "l")
		//
		windowsFlag = strings.Contains(content, windowsSeparator)
		linuxFlag   = strings.Contains(content, linuxSeparator)
	)
	if windowsFlag && linuxFlag {
		applicationCrashed("Content contains both separators.")
		return separator
	}
	if strings.Contains(content, windowsSeparator) {
		return windowsSeparator
	}
	if strings.Contains(content, linuxSeparator) {
		return linuxSeparator
	}
	applicationCrashed("Content not contains separators.")
	return separator
}

func osTextTemplate(separator, os string) string {
	switch os {
	case "w":
		return fmt.Sprintf("\n**\n%s:\n", separator)
	case "l":
		return fmt.Sprintf("\r\n**\r\n%s:\r\n", separator)
	default:
		applicationCrashed("OS template isn't exist.")
		return separator
	}
}

func applicationCrashed(err interface{}) {
	log.Fatal(runtimeinfo.Runtime(2), ", ERROR: ", err)
}


func lemming(repositoriesFiles []string, group string, lemmatizer *golem.Lemmatizer, doClear textClearing.DoClear) {
	root := "group-by-clear/clear-by-stop-words-lemmas/" + group
	var getFile = func(repo string, group string) (string, string) {
		name := strings.Split(repo, ".txt")[0]
		f := name + "-" + group + ".txt"
		groupingFile := getDirPath(f, "group-by-elements/"+group)
		content, _ := ReadBytes(groupingFile)
		return content, f
	}
	for _, repo := range repositoriesFiles {
		content, fileName := getFile(repo, group)
		if group == "bodies" {
			textClearing.ClearMarkdown(&content)
		}
		sep := ""
		if strings.Contains(content, "\r\n") {
			sep = "\r\n"
		} else {
			if strings.Contains(content, "\n") {
				sep = "\n"
			}
		}
		elements := strings.Split(content, sep)
		writeContent := make([]string, 0)
		for _, element := range elements {
			clearingText, _, err := doClear(&element)
			if err != nil {
				fmt.Println(repo, " :", err)
				continue
			}
			writeContent = append(writeContent, *clearingText)
		}
		path := getDirPath(fileName, root)
		output := []byte(strings.Join(writeContent, "\n"))
		err := ioutil.WriteFile(path, output, 0644)
		if err != nil {
			panic(err)
		}
	}
}



func createAndWriteDictionary(dictionaryFile string, repositoriesFiles []string) {
	var (
		mx     = new(sync.Mutex)
		wg     = new(sync.WaitGroup)
		corpus = make([]*text_preprocessing.VectorizedCorpusModel, 0)
	)
	for _, repo := range repositoriesFiles {
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
	path := getDirPath(dictionaryFile, "tests")
	output := []byte(strings.Join(writeContent, "\n"))
	err := ioutil.WriteFile(path, output, 0644)
	if err != nil {
		panic(err)
	}
}

func writeDictionaryAndVectorizedCorpus(vectorsFile, dictionaryFile string, repositoriesFiles []string) {
	path := getDirPath(dictionaryFile, "tests")
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
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
	for _, repo := range repositoriesFiles {
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
				if i > 5000 {

				}
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
	model := text_preprocessing.VectorizedWithDictionary(
		&dictionary,
		corpus...,
	)
	writeContent := make([][]string, 0)
	for item := range model.GetPresenceVectors().IterBuffered() {
		wg.Add(1)
		go func(item cmap.Tuple, writeContent *[][]string, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			val := item.Val.(*[]float64)
			content := make([]string, len(*val)+2)
			key := item.Key
			content[0] = key
			for i := 1; i < len(*val); i++ {
				v := strconv.Itoa(int((*val)[i-1]))
				content[i] = v
			}
			mx.Lock()
			*writeContent = append(*writeContent, content)
			mx.Unlock()
			return
		}(item, &writeContent, wg, mx)
	}
	wg.Wait()
	path = getDirPath(vectorsFile, "tests")
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		panic(err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, value := range writeContent {
		err := writer.Write(value)
		if err != nil {
			fmt.Println(err)
			panic(err)
			return
		}
	}
	fmt.Println("Finish...")
}
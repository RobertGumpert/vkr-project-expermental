package main

import (
	"fmt"
	"go-agregator/pckg/runtimeinfo"
	issuesCollector "go-agregator/pckg/scratching/github-api/issues-collector"
	textPreprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"io/ioutil"
	"log"
	"path/filepath"
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
	BodiesPreprocessing *textPreprocessing.TextPreprocessor
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
		descriptionPreprocessing *textPreprocessing.TextPreprocessor
		titlesPreprocessing      *textPreprocessing.TextPreprocessor
		bodiesPreprocessing      *textPreprocessing.TextPreprocessor
		do = func(content *string, preprocessor *textPreprocessing.TextPreprocessor, wg *sync.WaitGroup) {
			defer wg.Done()
			*preprocessor = *textPreprocessing.NewTextPreprocessor(*content).DOPullThread(5)
			return
		}
	)
	if doPreprocessing {
		wg.Add(2)
		descriptionPreprocessing = new(textPreprocessing.TextPreprocessor)
		titlesPreprocessing = new(textPreprocessing.TextPreprocessor)
		//bodiesPreprocessing      = new(textPreprocessing.TextPreprocessor)
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

func getDirPath(file string) string {
	absPath, err := filepath.Abs(fmt.Sprintf("../go-agregator/data/tests/%s", file))
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

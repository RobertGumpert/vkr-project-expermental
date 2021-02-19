package main

import (
	"bufio"
	"fmt"
	"go-agregator/pckg/runtimeinfo"
	textPreprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
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
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	runtime.GOMAXPROCS(runtime.NumCPU())
	getFileContent("terminal.txt")
	fmt.Println("Final...")
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if scanner.Text() == "end" {
			os.Exit(1)
		}
	}
	return
}

func getFileContent(file string) (*textPreprocessing.TextPreprocessor, *textPreprocessing.TextPreprocessor, *textPreprocessing.TextPreprocessor) {
	var (
		wg = new(sync.WaitGroup)
		//
		content                        = readFile(file)
		description, topics, titles, _ = separateContent(content)
		descriptionTopicsContent       = strings.Join([]string{*description, *topics}, " ")
		//
		descriptionPreprocessing = new(textPreprocessing.TextPreprocessor)
		titlesPreprocessing      = new(textPreprocessing.TextPreprocessor)
		// bodiesPreprocessing      = new(textPreprocessing.TextPreprocessor)
		//
		doPreprocessing = func(content *string, preprocessor *textPreprocessing.TextPreprocessor, wg *sync.WaitGroup) {
			defer wg.Done()
			*preprocessor = *textPreprocessing.NewTextPreprocessor(*content).DOPullThread(5)
			return
		}
	)
	//
	wg.Add(2)
	//
	go doPreprocessing(&descriptionTopicsContent, descriptionPreprocessing, wg)
	go doPreprocessing(titles, titlesPreprocessing, wg)
	//
	wg.Wait()
	//
	return descriptionPreprocessing, titlesPreprocessing, nil
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

//func getDescription(description, descriptionSeparator, topicsSeparator *string, content string, wg *sync.WaitGroup) {
//	defer wg.Done()
//	str := strings.Split(
//		strings.Split(content, *descriptionSeparator)[1],
//		*topicsSeparator,
//	)[0]
//	description = &str
//}
//
//func getTopics(topics, topicsSeparator, titlesSeparator *string, content string, wg *sync.WaitGroup) {
//	defer wg.Done()
//	str := strings.Split(
//		strings.Split(content, *topicsSeparator)[1],
//		*titlesSeparator,
//	)[0]
//	topics = &str
//}
//
//func getTitles(titles, titlesSeparator, bodiesSeparator *string, content string, wg *sync.WaitGroup) {
//	defer wg.Done()
//	str := strings.Split(
//		strings.Split(content, *titlesSeparator)[1],
//		*bodiesSeparator,
//	)[0]
//	titles = &str
//}

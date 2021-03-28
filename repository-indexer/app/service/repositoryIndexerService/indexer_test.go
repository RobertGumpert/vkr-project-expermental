package repositoryIndexerService

import (
	"fmt"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textClearing"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"testing"
)

var (
	root          = "C:/VKR/vkr-project-expermental/go-agregator/data/group-by-elements/topics+descriptions"
	lemmatizer, _ = golem.New(en.New())
)

func createDataModels() []dataModel.RepositoryModel {
	var (
		models = make([]dataModel.RepositoryModel, 0)
	)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	for _, fileInfo := range files {
		if fileInfo.Name() == "results.txt" {
			continue
		}
		fileName := strings.Join([]string{root, fileInfo.Name()}, "/")
		str, err := ioutil.ReadFile(fileName)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
		split := strings.Split(string(str), "\n")
		//
		textClearing.ClearASCII(&split[0])
		textClearing.ClearSymbols(&split[0])
		textClearing.ClearSpecialWord(&split[0])
		slice := textClearing.GetLemmas(&split[0], false, lemmatizer)
		split[0] = strings.Join(*slice, " ")
		//
		textClearing.ClearASCII(&split[1])
		textClearing.ClearSymbols(&split[1])
		split[1] = strings.Join(*textClearing.GetLemmas(&split[1], false, lemmatizer), " ")
		//
		models = append(models, dataModel.RepositoryModel{
			Name:        strings.Split(fileInfo.Name(), ".")[0],
			Description: split[0],
			Topics:      strings.Split(split[1], " "),
		})
	}
	return models
}

func TestIndexingFlow(t *testing.T) {
	models := createDataModels()
	result, err := Indexing(models)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	log.Println("START PRINT DICTIONARY...")
	count := 0
	for item := range result.GetDictionary().IterBuffered() {
		log.Println(item.Key)
		count++
	}
	runtimeinfo.LogInfo("COUNT = ", count, ", LEN = ", len(models))
	log.Println("FINISH PRINT DICTIONARY.")
	log.Println("START PRINT NEAREST...")
	for _, repository := range result.GetNearestRepositories() {
		log.Println("MAIN: ", repository.GetRepositoryName())
		type kv struct {
			Key   string
			Value float64
		}
		var ss []kv
		for k, v := range repository.GetNearestRepositories() {
			ss = append(ss, kv{k, v})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value > ss[j].Value
		})
		for _, kv := range ss {
			if kv.Value < 0.4 {
				continue
			}
			log.Println(fmt.Sprintf("\t\t\t%s = %f", kv.Key, kv.Value))
		}
		//for nearest, distance := range repository.GetNearestRepositories() {
		//	if distance > 0.0 {
		//		log.Println("\t\t\t", nearest, " = ", distance)
		//	}
		//}
	}
	log.Println("FINISH PRINT NEAREST.")
}

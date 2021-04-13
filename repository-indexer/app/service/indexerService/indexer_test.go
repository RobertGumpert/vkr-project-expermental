package indexerService

import (
	"fmt"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
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


func connect() repository.IRepository {
	sqlRepository := repository.NewSQLRepository(
		storageProvider,
	)
	return sqlRepository
}

var (
	storageProvider = repository.SQLCreateConnection(
		repository.TypeStoragePostgres,
		repository.DSNPostgres,
		nil,
		"postgres",
		"toster123",
		"vkr-db",
		"5432",
		"disable",
	)
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
	for i, fileInfo := range files {
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
		model := dataModel.RepositoryModel{
			Description: split[0],
			Topics:      strings.Split(split[1], " "),
		}
		model.ID = uint(i)
		models = append(models, model)

	}
	return models
}

func createRealData() []dataModel.RepositoryModel {
	db := connect()
	models, err := db.GetAllRepositories()
	if err != nil {
		log.Fatal(err)
	}
	return models
}

func TestIndexingFlow(t *testing.T) {
	models := createRealData()
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
	log.Println("COUNT = ", count, ", LEN = ", len(models))
	log.Println("FINISH PRINT DICTIONARY.")
	log.Println("START PRINT NEAREST...")
	for _, r := range result.GetNearestRepositories() {
		log.Println("MAIN: ", r.GetRepositoryID())
		type kv struct {
			Key   uint
			Value float64
		}
		var ss []kv
		for k, v := range r.GetNearestRepositories() {
			ss = append(ss, kv{k, v})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value > ss[j].Value
		})
		for _, kv := range ss {
			if kv.Value < 0.4 {
				continue
			}
			log.Println(fmt.Sprintf("\t\t\t%d = %f", kv.Key, kv.Value))
		}
	}
	log.Println("FINISH PRINT NEAREST.")
}

func TestIndexingIDFFlow(t *testing.T) {
	models := createDataModels()
	result, err := IndexingIDF(models, 3)
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
		log.Println("MAIN: ", repository.GetRepositoryID())
		type kv struct {
			Key   uint
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
			log.Println(fmt.Sprintf("\t\t\t%d = %f", kv.Key, kv.Value))
		}
	}
	log.Println("FINISH PRINT NEAREST.")
}
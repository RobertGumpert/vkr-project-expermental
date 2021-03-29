package hashRepository

import (
	"fmt"
	"github.com/RobertGumpert/gosimstor"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"gorm.io/gorm"
	"log"
	"strconv"
	"testing"
)

var (
	hash repository.IRepository
	root = "C:/VKR/vkr-project-expermental/repository-indexer"
)

func createRepository() {
	repo, err := NewLocalHashStorage(root)
	if err != nil {
		log.Fatal(err)
	}
	hash = repo
}

func TestInsertFlow(t *testing.T) {
	createRepository()
	count := 10
	var (
		keyWords = make([]gosimstor.Row, 0)
		nearests = make([]gosimstor.Row, 0)
	)
	for i := 0; i < count-1; i++ {
		//
		keyWord := gosimstor.Row{
			ID:   "Key-" + strconv.Itoa(i),
			Data: int64(i),
		}
		n := dataModel.RepositoryModel{}
		n.ID = uint(i) * 10
		nearest := gosimstor.Row{
			ID: uint(i),
			Data: dataModel.NearestRepositoriesJSON{Repositories: []dataModel.RepositoryModel{
				n,
			}},
		}
		//
		_, err := hash.AddKeyWord(keyWord.ID.(string), keyWord.Data.(int64), dataModel.RepositoriesIncludeKeyWordsJSON{})
		if err != nil {
			t.Fatal(err)
		}
		err = hash.AddNearestRepositories(nearest.ID.(uint), nearest.Data.(dataModel.NearestRepositoriesJSON))
		if err != nil {
			t.Fatal(err)
		}
		//
		keyWords = append(keyWords, keyWord)
		nearests = append(nearests, nearest)
	}
}

func TestReadFlow(t *testing.T) {
	createRepository()
	count := 10
	for i := 0; i < count; i++ {
		keyWord, err := hash.GetKeyWord("Key-" + strconv.Itoa(i))
		if err != nil {
			log.Println(err)
		}
		nearest, err := hash.GetNearestRepositories(uint(i))
		if err != nil {
			log.Println(err)
		}
		//
		log.Println(fmt.Sprintf("KEYWORD: id = '%s', data = '%d'", keyWord.KeyWord, keyWord.Position))
		log.Println(fmt.Sprintf("NEAREST: id = '%d'", uint(i)))
		for _, r := range nearest.Repositories {
			log.Println(fmt.Sprintf("\t\t\tNEAREST: data -> '%d'", r.ID))
		}
	}
}

func TestUpdateFlow(t *testing.T) {
	createRepository()
	_, err := hash.UpdateKeyWord("Key-"+strconv.Itoa(0), 10, dataModel.RepositoriesIncludeKeyWordsJSON{})
	if err != nil {
		t.Fatal(err)
	}
	keyWord, err := hash.GetKeyWord("Key-" + strconv.Itoa(0))
	if err != nil {
		t.Fatal(err)
	}
	log.Println(fmt.Sprintf("KEYWORD: id = '%s', data = '%d'", keyWord.KeyWord, keyWord.Position))
	err = hash.UpdateNearestRepositories(0, dataModel.NearestRepositoriesJSON{Repositories: []dataModel.RepositoryModel{
		{
			Model: gorm.Model{ID: 80},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	nearest, err := hash.GetNearestRepositories(0)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(fmt.Sprintf("NEAREST: id = '%d'", 0))
	for _, r := range nearest.Repositories {
		log.Println(fmt.Sprintf("\t\t\tNEAREST: data -> '%d'", r.ID))
	}
	err = hash.UpdateNearestRepositories(1, dataModel.NearestRepositoriesJSON{Repositories: []dataModel.RepositoryModel{
		{
			Model: gorm.Model{ID: 90},
		},
		{
			Model: gorm.Model{ID: 100},
		},
		{
			Model: gorm.Model{ID: 110},
		},
		{
			Model: gorm.Model{ID: 120},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	nearest, err = hash.GetNearestRepositories(1)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(fmt.Sprintf("NEAREST: id = '%d'", 1))
	for _, r := range nearest.Repositories {
		log.Println(fmt.Sprintf("\t\t\tNEAREST: data -> '%d'", r.ID))
	}
}

func TestRewriteFlow(t *testing.T) {
	createRepository()
	count := 10
	var (
		keyWords     = make([]dataModel.RepositoriesKeyWordsModel, 0)
		nearests     = make([]dataModel.NearestRepositoriesJSON, 0)
		repositories = make([]uint, 0)
	)
	for i := 0; i < count; i++ {
		//
		keyWords = append(keyWords, dataModel.RepositoriesKeyWordsModel{
			Model:        gorm.Model{},
			KeyWord:      "Key-" + strconv.Itoa(i),
			Position:     int64(i * 10),
			Repositories: nil,
		})
		nearests = append(nearests, dataModel.NearestRepositoriesJSON{Repositories: []dataModel.RepositoryModel{
			{
				Model: gorm.Model{ID: uint(i) * 100},
			},
			{
				Model: gorm.Model{ID: uint(i) * 100 + 10},
			},
		}})
		repositories = append(repositories, uint(i))
	}

	err := hash.RewriteAllKeyWords(keyWords)
	if err != nil {
		t.Fatal(err)
	}
	err = hash.RewriteAllNearestRepositories(repositories, nearests)
	if err != nil {
		t.Fatal(err)
	}
}

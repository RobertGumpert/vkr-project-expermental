package issuesComparator

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
	"log"
	"runtime"
	"strings"
	"testing"
)

func customGettingResultFunction(resultCompare interface{}) {
	runtimeinfo.LogInfo("RESULT RECEIVED:",
		" cosine :[",
		resultCompare.(dataModel.NearestIssues).CosineDistance,
		"]",
		"main issue: [", resultCompare.(dataModel.NearestIssues).IssueID, "],",
		" second issue: [", resultCompare.(dataModel.NearestIssues).NearestIssueID, "]",
	)
}

func readTitlesFromFiles() map[string][]dataModel.Issue {
	p := "C:/VKR/vkr-project-expermental/go-agregator/data/group-by-elements/titles"
	fileNames, err := ioutil.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	files := make(map[string][]dataModel.Issue)
	for i, fileName := range fileNames {
		dataModels := make([]dataModel.Issue, 0)
		filePath := fmt.Sprintf("%s/%s", p, fileName.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		var titles []string
		text := string(content)
		if strings.Contains(text, "\r\n") {
			titles = strings.Split(text, "\r\n")
		}
		if strings.Contains(text, "\n") {
			titles = strings.Split(text, "\n")
		}
		for j, title := range titles {
			dataModels = append(
				dataModels,
				dataModel.Issue{
					RepositoryID: uint(i),
					Number:       j,
					Title:        title,
				},
			)
		}
		key := strings.Split(fileName.Name(), "-titles.txt")[0]
		files[key] = dataModels
	}
	return files
}

func TestPairRepositoryFlow(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	files := readTitlesFromFiles()
	comparator := NewComparator(
		1000,
		100,
		70,
		customGettingResultFunction,
	)
	resultChannel := comparator.AddCompareIssuesInPairs(
		files["react"],
		files["vue"],
		comparator.CompareOnlyTitles,
	)
	for result := range resultChannel {
		log.Println("RESULT: ", result)
		break
	}
	log.Println("OK!")
}

func createFakeFrequencyJSON(str string) ([]byte, error) {
	slice := textDictionary.TextTransformToFeaturesSlice(str)
	dict := textVectorized.GetFrequencyMap(slice)
	m := make(map[string]float64, 0)
	for item := range dict.IterBuffered() {
		m[item.Key] = item.Val.(float64)
	}
	obj := &dataModel.TitleFrequencyJSON{Dictionary: m}
	return json.Marshal(obj)
}

func createFakeRepositories() ([]dataModel.Issue, []dataModel.Issue) {
	titleA := "Feature Request: Feature Request Warnings for missing Aria properties in debug mode"
	// missing Aria properties in debug mode
	titleB := "Feature Request: "
	//
	btsA, err := createFakeFrequencyJSON(titleA)
	btsB, err := createFakeFrequencyJSON(titleB)
	if err != nil {
		panic(err)
	}
	//
	main := make([]dataModel.Issue, 0)
	second := make([]dataModel.Issue, 0)
	for i := 0; i < 17; i++ {
		main = append(main, dataModel.Issue{
			Model: gorm.Model{
				ID: uint(i),
			},
			RepositoryID:       1,
			Title:              titleA,
			TitleFrequencyJSON: btsA,
		})
		second = append(second, dataModel.Issue{
			Model: gorm.Model{
				ID: uint(i),
			},
			RepositoryID:       2,
			Title:              titleB,
			TitleFrequencyJSON: btsB,
		})
	}
	return main, second
}

func TestFakeRepositoriesFlow(t *testing.T) {
	main, second := createFakeRepositories()
	comparator := NewComparator(
		1000,
		100,
		70,
		customGettingResultFunction,
	)
	resultChannel := comparator.AddCompareIssuesInPairs(
		main,
		second,
		comparator.CompareOnlyTitles,
	)
	for result := range resultChannel {
		log.Println("RESULT: ", result)
		break
	}
	log.Println("OK!")
	//
	resultChannel = comparator.AddCompareIssuesInPairs(
		main,
		second,
		comparator.CompareOnlyTitlesWithDictionaries,
	)
	for result := range resultChannel {
		log.Println("RESULT: ", result)
		break
	}
}

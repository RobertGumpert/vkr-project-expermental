package issuesComparator

import (
	"fmt"
	"io/ioutil"
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
	"log"
	"runtime"
	"strings"
	"testing"
)

func customGettingResultFunction(resultCompare interface{}) {
	runtimeinfo.LogInfo("RESULT RECEIVED: ", resultCompare.(dataModel.NearestIssues).CosineDistance)
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
					RepositoryID:  uint(i),
					Number:        j,
					URL:           "",
					Title:         title,
					State:         "",
					Body:          "",
					NearestIssues: nil,
					TurnIn:        nil,
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

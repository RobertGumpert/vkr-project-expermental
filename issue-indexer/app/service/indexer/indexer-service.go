package indexer

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
)

type Indexer struct {
	MaxCountThreads int
}

func NewIndexer(maxCountThreads int) *Indexer {
	return &Indexer{MaxCountThreads: maxCountThreads}
}

func (indexer *Indexer) Do(main, second []dataModel.Issue) {
	countThreads := len(main) / indexer.MaxCountThreads
}

func (indexer *Indexer) do(main, second []dataModel.Issue) {
	for i := 0; i < len(main); i++ {
		for j := 0; j < len(second); j++ {
			corpus := make([]string, 2)
			corpus[0] = main[i].Title
			corpus[1] = second[j].Title
			dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(
				corpus,
				textPreprocessing.LinearMode,
			)
			if countFeatures == 0 {
				continue
			}
			bagOfWords := textVectorized.FrequencyVectorized(
				vectorsOfWords,
				dictionary,
				textPreprocessing.LinearMode,
			)
		}
	}
}

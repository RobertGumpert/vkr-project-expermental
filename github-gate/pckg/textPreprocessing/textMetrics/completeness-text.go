package textMetrics

import (
	"issue-indexer/pckg/textPreprocessing"
	"sync"
)


func CompletenessText(bagOfWords [][]float64, mode textPreprocessing.ThreadMode) []float64 {
	if mode == textPreprocessing.ParallelMode {
		return parallelCalculateCompletenessText(bagOfWords)
	}
	return linearCalculateCompletenessText(bagOfWords)
}

func linearCalculateCompletenessText(bagOfWords [][]float64) []float64 {
	var (
		completenessMatrix = make([]float64, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		completeness := calculateCompletenessText(bagOfWords[i])
		completenessMatrix[i] = completeness
	}
	return completenessMatrix
}

func parallelCalculateCompletenessText(bagOfWords [][]float64) []float64 {
	var (
		wg                 = new(sync.WaitGroup)
		completenessMatrix = make([]float64, len(bagOfWords))
	)
	for i := 0; i < len(bagOfWords); i++ {
		wg.Add(1)
		go func(i int, bagOfWords [][]float64, completenessMatrix []float64, wg *sync.WaitGroup) {
			defer wg.Done()
			completeness := calculateCompletenessText(bagOfWords[i])
			completenessMatrix[i] = completeness
			return
		}(i, bagOfWords, completenessMatrix, wg)
	}
	wg.Wait()
	return completenessMatrix
}

func calculateCompletenessText(vector []float64) float64 {
	var(
		completeness   = float64(0)
	)
	for i := 0; i < len(vector); i++ {
		vectorI := switchToFloat64(vector[i])
		if vectorI > 0 {
			completeness++
		}
	}
	completeness = (completeness / float64(len(vector))) * 100
	return completeness
}


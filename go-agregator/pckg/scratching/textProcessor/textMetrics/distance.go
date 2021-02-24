package textMetrics

import (
	"errors"
	cmap "github.com/streamrail/concurrent-map"
	"math"
)

func CosineDistance(vecA, vecB *cmap.ConcurrentMap) (float64, error) {
	if vecA.Count() != vecB.Count() {
		return -1, errors.New("Vectors must be equals size. ")
	}
	numerator, denumerator := float64(0), float64(0)
	squareA, squareB := float64(0), float64(0)
	for item := range vecA.IterBuffered() {
		if valB, exist := vecB.Get(item.Key); !exist {
			return -1, errors.New("Vectors must be built using one dictionary. ")
		} else {
			a := switchToFloat64(item.Val)
			b := switchToFloat64(valB)
			numerator += a * b
			squareA += a * a
			squareB += b * b
		}
	}
	squareA = math.Sqrt(squareA)
	squareB = math.Sqrt(squareB)
	denumerator = squareA * squareB
	return numerator / denumerator, nil
}

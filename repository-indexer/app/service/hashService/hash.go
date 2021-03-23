package hashService

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	concurrentMap "github.com/streamrail/concurrent-map"
)

type HashStorageService struct {
	dictionary concurrentMap.ConcurrentMap
	links      concurrentMap.ConcurrentMap
	bagOfWords [][]float64
	db         repository.IRepository
	minIdf     uint
}

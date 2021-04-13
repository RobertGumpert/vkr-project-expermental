package appService

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	concurrentMap "github.com/streamrail/concurrent-map"
	"repository-indexer/app/service/indexerService"
)

func (service *AppService) reindexing() (
	repositories []uint,
	nearests []dataModel.NearestRepositoriesJSON,
	keyWords []dataModel.RepositoriesKeyWordsModel,
	err error,
) {
	var (
		models     []dataModel.RepositoryModel
		dictionary concurrentMap.ConcurrentMap
	)
	keyWords = make([]dataModel.RepositoriesKeyWordsModel, 0)
	repositories = make([]uint, 0)
	nearests = make([]dataModel.NearestRepositoriesJSON, 0)
	//
	models, err = service.mainStorage.GetAllRepositories()
	if err != nil {
		return repositories, nearests, keyWords, err
	}
	results, err := indexerService.Indexing(models)
	if err != nil {
		return repositories, nearests, keyWords, err
	}
	dictionary = results.GetDictionary()
	for item := range dictionary.IterBuffered() {
		keyWord := item.Key
		position := item.Val.(int64)
		keyWords = append(
			keyWords,
			dataModel.RepositoriesKeyWordsModel{
				KeyWord:  keyWord,
				Position: position,
			},
		)
	}
	err = service.localStorage.RewriteAllKeyWords(keyWords)
	if err != nil {
		return repositories, nearests, keyWords, err
	}
	for _, repositoryAnalise := range results.GetNearestRepositories() {
		repositoryID := repositoryAnalise.GetRepositoryID()
		nearest := repositoryAnalise.GetNearestRepositories()
		repositories = append(
			repositories,
			repositoryID,
		)
		mp := make(map[uint]float64)
		for id, distance := range nearest {
			distance = distance * 100
			if distance >= service.config.MinimumCosineDistanceThreshold {
				mp[id] = distance
			}
		}
		nearests = append(
			nearests,
			dataModel.NearestRepositoriesJSON{
				Repositories: mp,
			},
		)
	}
	err = service.localStorage.RewriteAllNearestRepositories(repositories, nearests)
	if err != nil {
		return repositories, nearests, keyWords, err
	}
	return repositories, nearests, keyWords, nil
}

func (service *AppService) getIndexerForRepository(settings *jsonSendFromGateReindexingForRepository) doReindexing {
	return func() {
		result := resultIndexing{
			taskType: taskTypeReindexingForRepository,
			taskKey:  settings.TaskKey,
			jsonBody: jsonSendToGateReindexingForRepository{
				ExecutionTaskStatus: jsonExecutionTaskStatus{
					TaskKey:       settings.TaskKey,
					TaskCompleted: true,
				},
				Result: jsonNearestRepository{},
			},
		}
		repositories, nearests, _, err := service.reindexing()
		if err != nil {
			jsonBody := result.jsonBody.(jsonSendToGateReindexingForRepository)
			jsonBody.ExecutionTaskStatus.Error = err
			result.jsonBody = jsonBody
			service.chanResult <- result
			return
		}
		jsonBody := result.jsonBody.(jsonSendToGateReindexingForRepository)
		var jsonResultModel jsonNearestRepository
		for next := 0; next < len(repositories); next++ {
			if repositories[next] == settings.RepositoryID {
				jsonResultModel = jsonNearestRepository{
					RepositoryID:          repositories[next],
					NearestRepositoriesID: nearests[next].Repositories,
				}
				jsonBody.Result = jsonResultModel
				break
			}
		}
		result.jsonBody = jsonBody
		service.chanResult <- result
		return
	}
}

func (service *AppService) getIndexerForAll(settings *jsonSendFromGateReindexingForAll) doReindexing {
	return func() {

		result := resultIndexing{
			taskType: taskTypeReindexingForAll,
			taskKey:  settings.TaskKey,
			jsonBody: jsonSendToGateReindexingForAll{
				ExecutionTaskStatus: jsonExecutionTaskStatus{
					TaskKey:       settings.TaskKey,
					TaskCompleted: true,
				},
				Results: make([]jsonNearestRepository, 0),
			},
		}
		repositories, nearests, _, err := service.reindexing()
		if err != nil {
			jsonBody := result.jsonBody.(jsonSendToGateReindexingForAll)
			jsonBody.ExecutionTaskStatus.Error = err
			result.jsonBody = jsonBody
			service.chanResult <- result
			return
		} else {
			jsonBody := result.jsonBody.(jsonSendToGateReindexingForAll)
			jsonResultModels := make([]jsonNearestRepository, 0)
			for next := 0; next < len(repositories); next++ {
				jsonResultModels = append(
					jsonResultModels,
					jsonNearestRepository{
						RepositoryID:          repositories[next],
						NearestRepositoriesID: nearests[next].Repositories,
					},
				)
			}
			jsonBody.Results = jsonResultModels
			result.jsonBody = jsonBody
		}
		service.chanResult <- result
		return
	}
}

func (service *AppService) getIndexerForGroupRepositories(settings *jsonSendFromGateReindexingForGroupRepositories) doReindexing {
	return func() {
		ids := func() map[uint]struct{} {
			mp := make(map[uint]struct{})
			for _, id := range settings.RepositoryID {
				mp[id] = struct{}{}
			}
			return mp
		}()
		result := resultIndexing{
			taskType: taskTypeReindexingForGroupRepositories,
			taskKey:  settings.TaskKey,
			jsonBody: jsonSendToGateReindexingForGroupRepositories{
				ExecutionTaskStatus: jsonExecutionTaskStatus{
					TaskKey:       settings.TaskKey,
					TaskCompleted: true,
				},
				Results: make([]jsonNearestRepository, 0),
			},
		}
		repositories, nearests, _, err := service.reindexing()
		if err != nil {
			jsonBody := result.jsonBody.(jsonSendToGateReindexingForGroupRepositories)
			jsonBody.ExecutionTaskStatus.Error = err
			result.jsonBody = jsonBody
			service.chanResult <- result
			return
		} else {
			jsonBody := result.jsonBody.(jsonSendToGateReindexingForGroupRepositories)
			jsonResultModels := make([]jsonNearestRepository, 0)
			for next := 0; next < len(repositories); next++ {
				if _, exist := ids[repositories[next]]; exist {
					jsonResultModels = append(
						jsonResultModels,
						jsonNearestRepository{
							RepositoryID:          repositories[next],
							NearestRepositoriesID: nearests[next].Repositories,
						},
					)
				}
			}
			jsonBody.Results = jsonResultModels
			result.jsonBody = jsonBody
		}
		service.chanResult <- result
		return
	}
}

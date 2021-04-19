package hashRepository

import (
	"github.com/RobertGumpert/gosimstor"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"sort"
)

//
// IMPLEMENT
//

func (storage *LocalHashStorage) CloseConnection() error {
	return gosimstor.Destructor(storage.storage)
}

func (storage *LocalHashStorage) AddKeyWord(keyWord string, position int64, repositories dataModel.RepositoriesIncludeKeyWordsJSON) (dataModel.RepositoriesKeyWordsModel, error) {
	var (
		model dataModel.RepositoriesKeyWordsModel
	)
	return model, storage.storage.Insert(
		dictionary,
		gosimstor.Row{
			ID:   keyWord,
			Data: position,
		},
	)
}

func (storage *LocalHashStorage) UpdateKeyWord(keyWord string, position int64, repositories dataModel.RepositoriesIncludeKeyWordsJSON) (dataModel.RepositoriesKeyWordsModel, error) {
	var (
		model dataModel.RepositoriesKeyWordsModel
	)
	return model, storage.storage.Update(
		dictionary,
		gosimstor.Row{
			ID:   keyWord,
			Data: position,
		},
	)
}

func (storage *LocalHashStorage) RewriteAllKeyWords(models []dataModel.RepositoriesKeyWordsModel) error {
	var (
		rows = make([]gosimstor.Row, 0)
	)
	for _, model := range models {
		rows = append(rows, gosimstor.Row{
			ID:   model.KeyWord,
			Data: model.Position,
		})
	}
	return storage.storage.Rewrite(
		dictionary,
		rows,
	)
}

func (storage *LocalHashStorage) GetKeyWord(keyWord string) (dataModel.RepositoriesKeyWordsModel, error) {
	var (
		model = dataModel.RepositoriesKeyWordsModel{}
		err   error
	)
	row, err := storage.storage.Read(
		dictionary,
		keyWord,
	)
	if err != nil {
		return model, err
	}
	model.KeyWord = row.ID.(string)
	model.Position = row.Data.(int64)
	return model, err
}

func (storage *LocalHashStorage) GetAllKeyWords() ([]dataModel.RepositoriesKeyWordsModel, error) {
	var (
		model = make([]dataModel.RepositoriesKeyWordsModel, 0)
		err   error
	)
	ids, err := storage.storage.GetIDs(
		dictionary,
	)
	if err != nil {
		return model, err
	}
	for _, id := range ids {
		row, err := storage.storage.Read(
			dictionary,
			id,
		)
		if err != nil {
			continue
		}
		model = append(model, dataModel.RepositoriesKeyWordsModel{KeyWord: row.ID.(string), Position: row.Data.(int64)})
	}
	return model, err
}

func (storage *LocalHashStorage) AddNearestRepositories(repositoryId uint, nearest dataModel.NearestRepositoriesJSON) error {
	return storage.storage.Insert(
		nearestRepositories,
		gosimstor.Row{
			ID:   repositoryId,
			Data: nearest,
		},
	)
}

func (storage *LocalHashStorage) UpdateNearestRepositories(repositoryId uint, nearest dataModel.NearestRepositoriesJSON) error {
	return storage.storage.Update(
		nearestRepositories,
		gosimstor.Row{
			ID:   repositoryId,
			Data: nearest,
		},
	)
}

func (storage *LocalHashStorage) GetNearestRepositories(repositoryId uint) (dataModel.NearestRepositoriesJSON, error) {
	var (
		model dataModel.NearestRepositoriesJSON
		err   error
	)
	row, err := storage.storage.Read(
		nearestRepositories,
		repositoryId,
	)
	if err != nil {
		return model, err
	}
	model = row.Data.(dataModel.NearestRepositoriesJSON)
	model.Repositories = rankByDistance(model.Repositories)
	return model, err
}

func (storage *LocalHashStorage) RewriteAllNearestRepositories(repositoryId []uint, models []dataModel.NearestRepositoriesJSON) error {
	var (
		rows = make([]gosimstor.Row, 0)
	)
	for i := 0; i < len(repositoryId); i++ {
		rows = append(rows, gosimstor.Row{
			ID:   repositoryId[i],
			Data: models[i],
		})
	}
	return storage.storage.Rewrite(
		nearestRepositories,
		rows,
	)
}

type tuple struct {
	Key   uint
	Value float64
}
type tupleList []tuple


func rankByDistance(wordFrequencies map[uint]float64) map[uint]float64 {
	pl := make(tupleList, 0)
	for k, v := range wordFrequencies {
		pl = append(pl, tuple{k, v})
	}
	sort.Slice(pl, func(i, j int) bool {
		return pl[i].Value > pl[j].Value
	})
	sortMp := make(map[uint]float64)
	for _, elem := range pl {
		sortMp[elem.Key] = elem.Value
	}
	return sortMp
}

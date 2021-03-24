package localStorage

import (
	"errors"
	concurrentMap "github.com/streamrail/concurrent-map"
	"runtime"
)

type FromStringToType func(data string) (interface{}, error)
type ToString func(data interface{}) (string, FromStringToType, error)

type Storage struct {
	fileStorage concurrentMap.ConcurrentMap
}

func NewStorage(constructors ...FileProviderConstructor) (*Storage, error) {
	storage := new(Storage)
	storage.fileStorage = concurrentMap.New()
	for i := 0; i < len(constructors); i++ {
		constructor := constructors[i]
		provider, err := constructor()
		if err != nil {
			return nil, err
		}
		storage.fileStorage.Set(provider.key, provider)
	}
	return storage, nil
}

func Destructor(storage *Storage) error {
	for item := range storage.fileStorage.IterBuffered() {
		dataModel := item.Val.(*fileProvider)
		mx := *dataModel.mx
		mx.Lock()
		err := dataModel.file.Close()
		if err != nil {
			return err
		}
		dataModel = nil
		mx.Unlock()
	}
	storage = nil
	runtime.GC()
	return nil
}

func (storage *Storage) WriteNew(providerKey string, id, data interface{}) error {
	var (
		provider *fileProvider
	)
	if inter, exist := storage.fileStorage.Get(providerKey); !exist {
		return errors.New("FILE PROVIDER ISN'T EXIST. ")
	} else {
		provider = inter.(*fileProvider)
	}
	return provider.WriteNew(
		id,
		data,
	)
}

func (storage *Storage) Read(providerKey string, id interface{}) (interface{}, interface{}, error) {
	var (
		provider *fileProvider
	)
	if inter, exist := storage.fileStorage.Get(providerKey); !exist {
		return nil, nil, errors.New("FILE PROVIDER ISN'T EXIST. ")
	} else {
		provider = inter.(*fileProvider)
	}
	return provider.Read(id)
}

func (storage *Storage) Update(providerKey string, id, data interface{}) error {
	var (
		provider *fileProvider
	)
	if inter, exist := storage.fileStorage.Get(providerKey); !exist {
		return errors.New("FILE PROVIDER ISN'T EXIST. ")
	} else {
		provider = inter.(*fileProvider)
	}
	return provider.Update(
		id,
		data,
	)
}

func (storage *Storage) Rewrite(providerKey string, id, data []interface{}) error {
	var (
		provider *fileProvider
	)
	if inter, exist := storage.fileStorage.Get(providerKey); !exist {
		return errors.New("FILE PROVIDER ISN'T EXIST. ")
	} else {
		provider = inter.(*fileProvider)
	}
	return provider.Rewrite(
		id,
		data,
	)
}
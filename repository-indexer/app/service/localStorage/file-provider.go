package localStorage

import (
	"bufio"
	"errors"
	concurrentMap "github.com/streamrail/concurrent-map"
	"os"
	"strings"
	"sync"
)

const (
	MAXLENGTH = 1000
)

type FileProviderConstructor func() (provider *fileProvider, e error)

type fileProvider struct {
	filePath, key      string
	file               *os.File
	mx                 *sync.Mutex
	pointers           concurrentMap.ConcurrentMap
	maxLengthIncrement int
	//
	convertData, convertID                     ToString
	convertDataFromString, convertIdFromString FromStringToType
}

func NewFileProvider(fileStorageName string, maxLengthIncrement int, convertID, convertData ToString) FileProviderConstructor {
	return func() (provider *fileProvider, e error) {
		return newFileProvider(fileStorageName, maxLengthIncrement, convertID, convertData)
	}
}

func newFileProvider(fileStorageName string, maxLengthIncrement int, convertID, convertData ToString) (*fileProvider, error) {
	provider := new(fileProvider)
	provider.mx = new(sync.Mutex)
	provider.pointers = concurrentMap.New()
	provider.key = fileStorageName
	provider.convertID = convertID
	provider.convertData = convertData
	provider.maxLengthIncrement = maxLengthIncrement
	if err := provider.openStorage(fileStorageName); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *fileProvider) openStorage(fileName string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	provider.filePath = strings.Join([]string{
		dir,
		"data",
		fileName + ".txt",
	}, "/")
	file, err := os.OpenFile(provider.filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	provider.file = file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		id := strings.Split(line, "=")[0]
		if !provider.pointers.Has(id) {
			provider.pointers.Set(id, provider.pointers.Count())
		}
	}
	return nil
}

func (provider *fileProvider) write(id, data interface{}) error {
	provider.mx.Lock()
	defer provider.mx.Unlock()
	var (
		pointer    = provider.pointers.Count()
		line, tail string
	)
	toStringID, convertIdFromString, err := provider.convertID(id)
	if err != nil {
		return err
	}
	toStringData, convertDataFromString, err := provider.convertData(data)
	if err != nil {
		return err
	}
	if provider.convertDataFromString == nil {
		provider.convertDataFromString = convertDataFromString
	}
	if provider.convertIdFromString == nil {
		provider.convertIdFromString = convertIdFromString
	}
	if provider.pointers.Has(toStringID) {
		return errors.New("KEY IS EXIST. ")
	}
	provider.pointers.Set(toStringID, pointer)
	line = strings.Join([]string{toStringID, toStringData}, "=")
	tailSize := provider.maxLengthIncrement * MAXLENGTH - len(line)
	if tailSize < 0 {
		return errors.New("EXCEEDED PERMISSIBLE LINE LENGTH. ")
	}
	for i := 0; i < tailSize; i++ {
		tail += "."
	}
	_, err = provider.file.WriteString(strings.Join([]string{line, tail}, "|") + "\n")
	return err
}

func (provider *fileProvider) read(id interface{}) (interface{}, error) {
	provider.mx.Lock()
	defer provider.mx.Unlock()
	var (
		toStringID, _, err = provider.convertID(id)
		inter, data        interface{}
		exist              bool
		buffer             = make([]byte, provider.maxLengthIncrement*MAXLENGTH-1)
		pointer            int
		line               string
	)
	if err != nil {
		return nil, err
	}
	inter, exist = provider.pointers.Get(toStringID)
	if !exist {
		return nil, errors.New("ELEMENT NOT EXIST BY KEY. ")
	}
	pointer = inter.(int)
	_, err = provider.file.Seek(int64(pointer*provider.maxLengthIncrement*MAXLENGTH), 0)
	if err != nil {
		return nil, err
	}
	_, err = provider.file.Read(buffer)
	if err != nil {
		return nil, err
	}
	line = string(buffer)
	line = strings.Split(line, "|")[0]
	if strings.TrimSpace(line) == "" {
		return nil, errors.New("LINE IS EMPTY. ")
	}
	if strings.Contains(line, "=") {
		d := strings.Split(line, "=")[1]
		if strings.TrimSpace(d) == "" {
			return nil, errors.New("DATA IS EMPTY. ")
		} else {
			data, err = provider.convertDataFromString(d)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, errors.New("LINE NOT CONTAINS SEPARATOR. ")
	}
	return data, nil
}

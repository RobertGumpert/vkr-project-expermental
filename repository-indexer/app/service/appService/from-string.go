package appService

import "strconv"

func FromStringToKeyWord(keyWord string) (interface{}, error) {
	return keyWord, nil
}

func FromStringToPositionKeyWord(position string) (interface{}, error) {
	return strconv.ParseInt(position, 10, 64)
}

package localStorage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func fromStringToString(data string) (interface{}, error) {
	return data, nil
}

func fromStringToFloat64Vector(data string) (interface{}, error) {
	var (
		split  = strings.Split(data, ",")
		vector = make([]float64, 0)
	)
	for i := 0; i < len(split); i++ {
		element, err := strconv.ParseFloat(split[i], 64)
		if err != nil {
			return nil, err
		}
		vector = append(vector, element)
	}
	return vector, nil
}

func ToStringString(data interface{}) (string, FromStringToType, error) {
	var (
		convert, ok = data.(string)
	)
	if !ok {
		return convert, fromStringToString, errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	return convert, fromStringToString, nil
}

func ToStringFloat64Vector(data interface{}) (string, FromStringToType, error) {
	var (
		convert    string
		elements   = make([]string, 0)
		vector, ok = data.([]float64)
	)
	if !ok {
		return convert, fromStringToFloat64Vector, errors.New("DOESN'T CONVERT 'FLOAT64 VECTOR' TO STRING")
	}
	for i := 0; i < len(vector); i++ {
		elements = append(
			elements,
			fmt.Sprintf("%f", vector[i]),
		)
	}
	convert = strings.Join(elements, ",")
	return convert, fromStringToFloat64Vector, nil
}

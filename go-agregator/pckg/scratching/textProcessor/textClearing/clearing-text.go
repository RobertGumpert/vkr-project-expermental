package textClearing

import (
	"errors"
	"fmt"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/bbalet/stopwords"
	"github.com/grokify/html-strip-tags-go"
	"github.com/russross/blackfriday/v2"
	_ "github.com/writeas/go-strip-markdown"
	"regexp"
	"strings"
)

const (
	CodeRegex        = `(?i)[\w\d]+[.](?i)[\w\d]+[(](?i)[\w\d]{0,}[)]`
	AsciiRegex       = `[[:^ascii:]]`
	SymbolsRegex     = `[]\d%:$"';[&*=<>}{)(?!/.,\-_^@]`
	SpecialWordRegex = `(?i)([_\-&*:;#<>@""''=/+\d~^%]{0,}\w+[_\-&*:;#<>@""''=/+\d~^%]{1,}\w{0,})`
	UrlRegex         = `(?i)https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`
)

type Contains func(str *string) error
type Clear func(str *string)

// -> must be panic()
//
func clearByRegex(regex string, str *string) {
	clearText := strings.ToLower(*str)
	// -> must be panic()
	regularExpression := regexp.MustCompile(regex)
	clearText = regularExpression.ReplaceAllString(clearText, " ")
	*str = clearText
}

func containsByRegex(regex string, str *string) error {
	match, err := regexp.MatchString(regex, *str)
	if err != nil {
		return err
	}
	if match {
		return errors.New("String contains element declared in regular expression. ")
	}
	return nil
}

//
//--------------------------------------Clear---------------------------------------------------------------------------
//


func ContainsCode(str *string) error {
	return containsByRegex(CodeRegex, str)
}

func ContainsASCII(str *string) error {
	return containsByRegex(AsciiRegex, str)
}

func ContainsSymbols(str *string) error {
	return containsByRegex(SymbolsRegex, str)
}

func ContainsSpecialWord(str *string) error {
	return containsByRegex(SpecialWordRegex, str)
}

//
//--------------------------------------Contains------------------------------------------------------------------------
//

func ClearByRegex(str *string, regex string) {
	clearByRegex(regex, str)
}

func ClearMarkdown(str *string) {
	bts := blackfriday.Run([]byte(*str))
	clearText := string(bts)
	clearText = strings.ReplaceAll(clearText, "\n", " !*! ")
	clearByRegex(`(?i)<code>(.*?)</code>`, &clearText)
	clearText = strip.StripTags(clearText)
	fmt.Println(clearText)
	clearText = strings.ReplaceAll(clearText, " !*! ", "\n")
	slice := ToSlice(&clearText)
	s := make([]string, 0)
	for _, w := range *slice {
		if !strings.Contains(w, "&") {
			w = strings.TrimSpace(w)
			s = append(s, w)
		}
	}
	*str = strings.Join(s, " ")
}

func ClearASCII(str *string) {
	clearByRegex(AsciiRegex, str)
}

func ClearCode(str *string) {
	clearByRegex(CodeRegex, str)
}

func ClearSymbols(str *string) {
	clearByRegex(SymbolsRegex, str)
}

func ClearSpecialWord(str *string) {
	clearByRegex(SpecialWordRegex, str)
}

func ClearSingleCharacters(str *string) error {
	if len(*str) == 0 {
		return errors.New("String is empty. ")
	}
	words := ToSlice(str)
	output := make([]string, 0)
	for i := 0; i < len(*words); i++ {
		if len((*words)[i]) != 1 {
			output = append(output, (*words)[i])
		}
	}
	clearText := strings.Join(output, " ")
	*str = clearText
	return nil
}

func ToSlice(str *string) *[]string {
	words := strings.Fields(*str)
	return &words
}

func getLemma(str *string, lemmatizers ...*golem.Lemmatizer) error {
	var lemmatizer *golem.Lemmatizer
	if len(lemmatizers) == 0 {
		lemmatizer, _ = golem.New(en.New())
	} else {
		lemmatizer = lemmatizers[0]
	}
	if exist := lemmatizer.InDict(*str); !exist {
		err := errors.New("Lemma for word '" + *str + "' isn't exist. ")
		// fmt.Println(err)
		return err
	}
	lemma := lemmatizer.Lemma(*str)
	*str = lemma
	return nil
}

func GetLemmas(str *string, deleteWordsWithoutLemma bool, lemmatizers ...*golem.Lemmatizer) *[]string {
	var lemmatizer *golem.Lemmatizer
	slice := ToSlice(str)
	if len(lemmatizers) == 0 {
		lemmatizer, _ = golem.New(en.New())
	} else {
		lemmatizer = lemmatizers[0]
	}
	lemmas := make([]string, 0)
	for _, word := range *slice {
		if err := getLemma(&word, lemmatizer); err != nil && deleteWordsWithoutLemma == true {
			continue
		}
		lemmas = append(lemmas, word)
	}
	return &lemmas
}

type DoClear func(str *string) (*string, *[]string, error)

func CustomClear(deleteWordsWithoutLemma bool, lemmatizer *golem.Lemmatizer, stopRegex []Contains, clearRegex []Clear) DoClear {
	return func(str *string) (*string, *[]string, error) {
		if len(*str) == 0 {
			return nil, nil, errors.New("String is empty. ")
		}
		if stopRegex != nil {
			for _, stop := range stopRegex {
				if err := stop(str); err != nil {
					return nil, nil, err
				}
			}
		}
		if clearRegex != nil {
			for _, clear := range clearRegex {
				clear(str)
			}
		}
		clearText := stopwords.CleanString(*str, "en", true)
		clearByRegex(UrlRegex, &clearText)
		slice := GetLemmas(&clearText, deleteWordsWithoutLemma, lemmatizer)
		clearText = strings.Join(*slice, " ")
		if err := ClearSingleCharacters(&clearText); err != nil {
			return nil, nil, err
		}
		slice = ToSlice(&clearText)
		if len(clearText) == 0 {
			return nil, nil, errors.New("Clear string is empty. ")
		}
		if len(*slice) == 1 {
			return nil, nil, errors.New("Clear string is with 1 element. ")
		}
		return &clearText, slice, nil
	}
}

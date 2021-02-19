# Пагинация по страницам ISSUE репозитория. Пакет "github-api/issues-collector".

### Функция Collect

```go
list, err := issuesCollector.Collect(
	"https://api.github.com/repos/facebook/react",
	"https://api.github.com/repos/microsoft/terminal", 
)
```

Для указанных репозиториев выполняетя запрос по адресу:
```
https://api.github.com/repos/facebook/react/issues?state=all
```
тем самым находит абсолютно все первые 30 issue, то есть закрытые и открытые issue.
У первого issue в списке находит поле number, которое по сути является кол-вом всех issue
в репозитории.

Найденное число number делится на 100, тем самым определется кол-во страниц для пагинации,
с кол-вом элементов на одной странице равным 100.

Далее выполняется пагинация по страницам с помощью пакета "github-api/pages-iterator".

Результат сохраняется в защищенную шард хэш-таблицу, где ключом является ссылка на репозиторий,
а значением является приватная структура, содержащая:
- массив структур Issue (из пакета github-api/mapper)
- поле URL
- поле LastPage, последняя полученная страница
- поле CountPages, кол-во всех странц issue в репозитории
```
key   [string] : "https://api.github.com/repos/facebook/react"
value [struct] : *issuePagesIterator
```

### Функция CustomizablePagesCollect

```go
list, err := issuesCollector.CustomizablePagesCollect(
	issuesCollector.NewConfiguration(
		issuesCollector.SetURL("https://api.github.com/repos/facebook/react"),
		issuesCollector.SetPage(0),
		issuesCollector.SetCountPagesAll("https://api.github.com/repos/facebook/react"),
	),
	issuesCollector.NewConfiguration(
		issuesCollector.SetURL("https://api.github.com/repos/microsoft/terminal"),
		issuesCollector.SetPage(0),
		issuesCollector.SetCountPages(1),
	), 
)
```

Функция выполняет ручную настройку пагинации. В качестве параметра, она принимает массив типов Constructor.

Тип Constructor - это производный тип пакета "github-api/issues-collector" представляеющий из себя функцию,
принимающую в качестве параметра массив типов Option и возвращающую приватную структуру *issuePagesIterator.

Тип Option - это производный тип пакета "github-api/issues-collector" представляеющий из себя функцию, выполняющую
настройку поля структуры *issuePagesIterator.

Тип issuePagesIterator - приватная структура пакета "github-api/issues-collector", содержащая поля:
- массив структур Issue (из пакета github-api/mapper)
- поле URL
- поле LastPage, последняя полученная страница
- поле CountPages, кол-во всех странц issue в репозитории

#### Доступные функции, возвращающие тип Constructor.

##### NewConfiguration(options ...Option)

- Параметры: массив типов Option
- Возвращает: тип Constructor

#### Доступные функции, возвращающие тип Option.

##### SetPage(page int64)

- Параметры:
  - page - с какой страницы начать пагинацию.
- Возвращает: тип Option.

##### SetURL(url string)

- Параметры:
  - url - устанавливает url репозитория.
- Возвращает: тип Option.

##### SetCountPages(countPages int64)

- Параметры:
  - countPages - указать до какой страницы выполнять пагинацию.
- Возвращает: тип Option.

##### SetCountPagesAll(url string)

Если заранее неизвестно кол-во страниц для пагинации, то данная опция выполнит
поиск кол-ва страниц аналогично тому, как это делает функция Collect.

- Параметры:
  - url - устанавливает url репозитория.
- Возвращает: тип Option.
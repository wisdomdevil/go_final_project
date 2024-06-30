package repo

import (
	"fmt"
	"time"
)

// структуры, методы и интерфейс для абстрагирования параметров поиска
type DateSearchParam struct {
	Date time.Time
}

func (dp *DateSearchParam) GetQueryData() *QueryData {
	return &QueryData{
		Param:     dp.Date.Format(timeTemplate),
		Condition: "WHERE date LIKE :search",
	}
}

type TextSearchParam struct {
	Text string
}

func (tp *TextSearchParam) GetQueryData() *QueryData {
	return &QueryData{
		Param:     fmt.Sprintf("%%%s%%", tp.Text),
		Condition: "WHERE title LIKE :search OR comment LIKE :search",
	}
}

type SearchQueryData interface {
	GetQueryData() *QueryData
}

func QueryDataFromString(search string) SearchQueryData {
	searchDate, err := time.Parse("02.01.2006", search)
	if err != nil {
		return &TextSearchParam{Text: search}
	} else {
		return &DateSearchParam{Date: searchDate}
	}
}

type QueryData struct {
	Param     string
	Condition string
}

// ---------------------------

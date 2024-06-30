package dateutil

import (
	"time"

	"github.com/wisdomdevil/go_final_project/internal/parser"
)

// NextDate calculates the next task date using the specified repeat rule and dates (now and date)
// NextDate разбирает правило повтора и применяет его к переданным датам
// и возвращает следующую дату в формате строки
// date и now должны передаваться уже разобранными и проверенными
// now — время от которого ищется ближайшая дата;
// date — исходное время в формате 20060102, от которого начинается отсчёт повторений;
// repeat — правило повторения в одном из форматов:
// d <число> - задача переносится на указанное число дней;
// y - задача выполняется ежегодно;
// w <через запятую от 1 до 7> - задача назначается в указанные дни недели,
// где 1 — понедельник, 7 — воскресенье;
// m <через запятую от 1 до 31,-1,-2> [через запятую от 1 до 12] -
// задача назначается в указанные дни месяца.
func NextDate(now time.Time, date time.Time, repeat string) (string, error) {
	pr, err := parser.ParseRepeat(now, date, repeat)
	if err != nil {
		return "", err
	}

	d, err := pr.GetNextDate(now, date)
	if err != nil {
		return "", err
	}
	return d.Format("20060102"), nil
}

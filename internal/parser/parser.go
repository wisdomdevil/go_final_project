package parser

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DRepeat хранит число правила d
type DRepeat struct {
	num int
}

// ParseDRepeat заполняет структуру DRepeat
// Сигнатура правила d: d <число> — задача переносится на указанное число дней.
// Максимально допустимое число равно 400
func ParseDRepeat(rule []string) (*DRepeat, error) {
	next, err := strconv.Atoi(rule[1])
	if err != nil {
		return nil, fmt.Errorf("Error in checking days in repeat rule, got '%s'", rule[1])
	}
	if next > 0 && next <= 400 {
		return &DRepeat{num: next}, nil
	}
	return nil, fmt.Errorf("expected number of days less than 400, got '%s'", rule[1])
}

// GetNextDate вычисляет следующую дату по правилу d
// d <число> — задача переносится на указанное число дней.
func (dr *DRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	result := date
	for {
		result = result.AddDate(0, 0, dr.num)
		if result.After(now) {
			return result, nil
		}
	}
}

// --------------------------------------------------------

type YRepeat struct {
}

// ParseYRepeat заполняет структуру YRepeat
// y — задача выполняется ежегодно. Этот параметр не требует дополнительных уточнений.
func ParseYRepeat(rule []string) (*YRepeat, error) {
	return &YRepeat{}, nil
}

// GetNextDate вычисляет следующую дату по правилу y
// При выполнении задачи дата перенесётся на год вперёд
func (yr *YRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	i := 1

	for {
		result := date.AddDate(i, 0, 0)
		if result.After(now) {
			return result, nil
		}
		i++
	}
}

// --------------------------------------------------------

// WRepeat хранит список чисел правила w
// Сигнатура: w <через запятую от 1 до 7> — задача назначается в указанные дни недели,
// где 1 — понедельник, 7 — воскресенье
type WRepeat struct {
	nums []int
}

// ParseWRepeat заполняет структуру WRepeat
func ParseWRepeat(rule []string) (*WRepeat, error) {
	// var week: days of week in rule repeat
	if len(rule) == 1 {
		return nil, fmt.Errorf("Error in w rule.")
	}

	weekDays := []int{}
	x := strings.Split(rule[1], ",")
	for i := 0; i < len(x); i++ {
		num, err := strconv.Atoi(x[i])
		if err != nil || num > 7 || num < 1 {
			return nil, fmt.Errorf("Can not parse days for repeat value.")
		}
		weekDays = append(weekDays, num)
	}
	return &WRepeat{nums: weekDays}, nil
}

// GetNextDate вычисляет следующую дату по правилу w
// Из списка выбирается ближайший день (по номеру дня в неделе)
func (wr *WRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	startdate := startDateForMWrule(now, date)

	todayWeekday := startdate.Weekday()

	sort.Ints(wr.nums) // сортируем, чтобы сразу взять тот день, что больше номером, чем сегодняшний

	numDay := int(todayWeekday)
	if numDay == 7 {
		numDay = 0
	}

	for _, n := range wr.nums {
		if n > numDay {
			result := startdate.AddDate(0, 0, n-numDay)
			return result, nil
		}
	}

	increment := 7 - int(startdate.Weekday())

	result := startdate.AddDate(0, 0, increment+wr.nums[0])

	return result, nil
}

// --------------------------------------------------------

// MRepeat хранит список дней и список месяцев правила m
type MRepeat struct {
	mDays   []int
	mMonths []int
}

// hasMonths определяет, есть ли в правиле m месяцы
func (mr *MRepeat) hasMonths() bool {
	return len(mr.mMonths) > 0
}

// ParseMRepeat first checks the the second part of repeat rule string for "m" rule.
// -1 and -2 are converted to the last day of the month and day before last.
// Then checks the the third part of repeat rule string for "m" rule.
// Сигнатура: m <через запятую от 1 до 31,-1,-2> [через запятую от 1 до 12]
// задача назначается в указанные дни месяца.
// При этом вторая последовательность чисел опциональна и указывает на определённые месяцы
func ParseMRepeat(rule []string, now time.Time, date time.Time) (*MRepeat, error) {
	// Сначала всегда рассматриваем сегодняшний месяц.
	// Смотрим со всеми днями, а уж если не подходят предложенные правилом дни (если все < сегодня),
	// то месяц берем следующий месяц.
	// Если в правиле 31ое число, а в рассматриваемом месяце 30 дней,
	// то проверяется, чтобы в следующем месяце был 31 день и рассматриваем уже его (следующий месяц)

	if len(rule) == 1 || len(rule) > 3 {
		return nil, fmt.Errorf("Error in m rule.")
	}

	// определяем, есть ли месяцы в правиле
	hasMonths := false

	if len(rule) == 3 {
		hasMonths = true
	}

	days := []int{}

	daysInRule := strings.Split(rule[1], ",") // daysInRule - это дни в правиле m

	// определим, от какой даты (now или date) вычислять nextdate
	// d, err := time.Parse("20060102", date)
	// if err != nil {
	// 	return nil, err
	// }
	startdate := startDateForMWrule(now, date)

	for _, day := range daysInRule {
		num, err := strconv.Atoi(day)
		if err != nil {
			return nil, fmt.Errorf("Error in checking days in repeat rule 'm', got '%s'", day)
		}
		if num >= 1 && num <= 31 {
			days = append(days, num)
		} else if num == -1 {
			// time.Date принимает значения вне их обычных диапазонов, то есть
			// значения нормализуются во время преобразования
			// Чтобы рассчитать количество дней текущего месяца (t), смотрим на день следующего месяца
			t := Date(startdate.Year(), int(startdate.Month()+1), 0)
			days = append(days, int(t.Day()))
		} else if num == -2 {
			// time.Date принимает значения вне их обычных диапазонов, то есть
			// значения нормализуются во время преобразования
			// Чтобы рассчитать количество дней текущего месяца (t), смотрим на день следующего месяца
			t := Date(startdate.Year(), int(startdate.Month()+1), 0)
			days = append(days, int(t.Day())-1)
		} else {
			return nil, fmt.Errorf("Error in checking days in repeat rule 'm', got '%s'", day)
		}

	}

	// checks the the third part of repeat rule string for "m" rule.
	months := []int{}

	if hasMonths {
		monthsInRule := strings.Split(rule[2], ",") // monthsInRule - это месяцы в правиле m

		for _, month := range monthsInRule {
			num, err := strconv.Atoi(month)
			if err != nil || num < 1 || num > 12 {
				return nil, fmt.Errorf("Error in checking days in repeat rule 'm', got '%s'", month)
			}
			months = append(months, num)
		}
	}
	return &MRepeat{mDays: days, mMonths: months}, nil
}

// GetNextDate вычисляет следующую дату по правилу m
// Из списка выбирается ближайший день (по номеру дня в месяце)
// если номера месяцев указаны, то выбираются дни конкретных месяцев
func (mr *MRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	startdate := startDateForMWrule(now, date)

	sort.Ints(mr.mDays)

	// ниже проверяем, что день startdate не является больше, чем последнее число из mDays
	// если же больше, то startmonth надо сделать следующим месяцем
	var nextDay time.Time

	if !mr.hasMonths() {
		for _, day := range mr.mDays {
			if day > int(startdate.Day()) {
				nextDay = startdate.AddDate(0, 0, day-int(startdate.Day()))
				if nextDay.Day() != day {
					nextDay = Date(startdate.Year(), int(startdate.Month())+1, day)
				}
				return nextDay, nil
			}
		}

		if nextDay == Date(0001, 1, 1) { // 0001-01-01 00:00:00 +0000 UTC нулевой вариант времени
			startdate = Date(int(startdate.Year()), int(startdate.Month())+1, 1)
			for _, day := range mr.mDays {
				if day >= int(startdate.Day()) {
					nextDay = startdate.AddDate(0, 0, day-int(startdate.Day()))
					return nextDay, nil
				}
			}
		}
	}

	if mr.hasMonths() {
		sort.Ints(mr.mMonths)

		nextDay = ruleMwithMonth(startdate, mr.mDays, mr.mMonths)
		return nextDay, nil
	}

	return time.Time{}, fmt.Errorf("Error in checking days and months in 'm' repeat rule")
}

// ----------------------------------------------------------------

type RepeatRule interface {
	GetNextDate(now time.Time, date time.Time) (time.Time, error)
}

// ----------------------------------------------------------------

// Date returns time type from the int types of year, month and day.
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// startDateForMWrule selects startdate from now and date: selects a later date
func startDateForMWrule(now time.Time, date time.Time) time.Time {
	if date.After(now) {
		return date
	}
	return now
}

// ruleMwithMonth gets sorted mDays and mMonths and returns nextDay
func ruleMwithMonth(startdate time.Time, mDays []int, mMonths []int) time.Time {
	var nextDay time.Time

	for _, month := range mMonths {
		if month == int(startdate.Month()) {
			startdate = Date(startdate.Year(), month, 1)

			// dayInMonth is number of days in the current month.
			t := Date(startdate.Year(), int(startdate.Month())+1, 0) // день до следующего месяца
			dayInMonth := t.Day()

			for _, day := range mDays {
				if day > int(startdate.Day()) && day <= dayInMonth {
					gotDay := Date(startdate.Year(), int(startdate.Month()), day)
					nextDay = gotDay
					return nextDay
				} else if day > int(startdate.Day()) && day > dayInMonth {
					startdate = Date(startdate.Year(), int(startdate.Month())+1, 1)
				}
			}
		} else if month > int(startdate.Month()) { // else сделан для того,
			// чтобы 1 число следующего месяца тоже учитывалось в поиске
			startdate = Date(startdate.Year(), month, 1)

			// dayInMonth is number of days in the current month.
			t := Date(startdate.Year(), int(startdate.Month())+1, 0) // день до следующего месяца
			dayInMonth := t.Day()

			for _, day := range mDays {
				if day >= int(startdate.Day()) && day <= dayInMonth {
					gotDay := Date(startdate.Year(), int(startdate.Month()), day)
					nextDay = gotDay
					return nextDay
				} else if day > int(startdate.Day()) && day > dayInMonth {
					startdate = Date(startdate.Year(), int(startdate.Month())+1, 1)
				}
			}
		}
	}
	return nextDay
}

// ParseRepeat в зависимости от входных данных мы возвращаем правило повтора (структуру нужного типа)
func ParseRepeat(now time.Time, date time.Time, repeat string) (RepeatRule, error) {
	if repeat == "" {
		return nil, fmt.Errorf("Expected repeat, got an empty string.")
	}

	rule := strings.Split(repeat, " ")

	var parsedRepeat RepeatRule
	var err error

	log.Printf("rule[0] before switch is: %v", rule[0])
	switch {
	case rule[0] == "y":
		parsedRepeat, err = ParseYRepeat(rule)
		if err != nil {
			return nil, err // чтобы тест не падал на строке log.Fatal(err), надо вместо этого возвращать ошибку
		}
	case rule[0] == "d":
		parsedRepeat, err = ParseDRepeat(rule)
		if err != nil {
			return nil, err
		}
	case rule[0] == "w":
		parsedRepeat, err = ParseWRepeat(rule)
		if err != nil {
			return nil, err
		}
	case rule[0] == "m":
		parsedRepeat, err = ParseMRepeat(rule, now, date)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unkown repeat identifier %s", rule[0])
	}

	return parsedRepeat, nil
}

package models

import (
	"fmt"
	"log"
	"time"

	"github.com/wisdomdevil/go_final_project/internal/dateutil"
)

type Task struct {
	ID      string `json:"id"`      // uint   `json:"id"`
	Date    string `json:"date"`    // дата задачи в формате 20060102
	Title   string `json:"title"`   // заголовок задачи
	Comment string `json:"comment"` // комментарий к задаче
	Repeat  string `json:"repeat"`  // правило повторения
}

// ValidateAndNormalizeDate checks the incoming data and sets the next date of the event.
func (t *Task) ValidateAndNormalizeDate() error {
	if t.Title == "" {
		err := fmt.Errorf("The title field is empty.")
		return err
	}
	now := time.Now().Truncate(24 * time.Hour)
	log.Printf("Today is %v", now)

	if t.Date == "" {
		t.Date = now.Format("20060102")
		log.Println("If t.Date is null.")
		return nil
	}

	if t.Date == "today" {
		t.Date = now.Format("20060102")
		log.Printf("Check if %v is equal 'today'", t.Date)
		return nil
	}

	date, err := time.Parse("20060102", t.Date)
	log.Printf("Date after parsing: %v", date)
	if err != nil {
		err := fmt.Errorf("The field date is wrong")
		return err
	}

	dt, err := time.Parse("20060102", t.Date)
	if err != nil {
		return err
	}

	if now.After(date) { // Если дата меньше сегодняшнего числа
		// если правило повторения не указано или равно пустой строке, подставляется сегодняшнее число
		if t.Repeat == "" {
			log.Printf("Repeat rule is empty.")
			t.Date = now.Format("20060102")
		} else {
			log.Printf("Repeat rule is not empty.")
			nextDate, err := dateutil.NextDate(now, dt, t.Repeat)
			if err != nil {
				log.Printf("Error in NextDate function: %v", err)
				return err
			}
			t.Date = nextDate
		}
	}

	log.Printf("Returning t.Date in TaskCreationRequest function  %v.", t.Date)
	fmt.Println("Error in ValidateAndNormalizeDate:", err)
	return nil
}

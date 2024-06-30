package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/wisdomdevil/go_final_project/internal/dateutil"
)

const (
	MarshallingError    = "error in marshalling JSON"
	UnMarshallingError  = "error in unmarshalling JSON"
	ResponseWriteError  = "error in writing data"
	ReadingError        = "error in reading data"
	InvalidIdError      = "invalid id"
	IdMissingError      = "id is missing"
	InvalidDateError    = "invalid date"
	InvalidNowDateError = "invalid now date"
	InvalidRepeatError  = "invalid repeat value"
	InternalServerError = "internal server error"
	ValidatingDateError = "error in validating date"
)

// repeatRulePattern checks if the reapeat rule starts with correct letter
var repeatRulePattern *regexp.Regexp = regexp.MustCompile(`^([mwd]\s\S.*|y$)`)

type signinRequest struct {
	Password string `json:"password"`
}

type signinResponse struct {
	Token string `json:"token"`
}

// Обработка ошибок для возврата ошибки в виде json.
type apiError struct {
	Error string `json:"error"`
}

func NewApiError(err error) apiError {
	return apiError{Error: err.Error()}
}

func (e apiError) ToJson() ([]byte, error) {
	res, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RenderApiErrorAndResponse(w http.ResponseWriter, err error, status int) {
	apiErr := NewApiError(err)
	errorJson, err := apiErr.ToJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, string(errorJson), status)
}

// ------------------------------------

// GetNextDay find next day for the task
func GetNextDay(w http.ResponseWriter, r *http.Request) {
	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	//check date
	//Тут спорно, так и не смог определиться, так-то клиент может какую-нибудь дичь закинуть и вернется ошибка, значит это BadRequest, но ошибка произошла на стороне сервера.
	_, err := strconv.Atoi(date)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidDateError), http.StatusBadRequest)
		return
	}
	//такая же история что и выше, оставил BadRequest
	dtParsed, err := time.Parse("20060102", date)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidDateError), http.StatusBadRequest)
		return
	}

	//check now
	_, err = strconv.Atoi(now)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidNowDateError), http.StatusBadRequest)
		return
	}

	dtNow, err := time.Parse("20060102", now)
	if err != nil {
		err := fmt.Errorf("wrong date: %v", err)
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidNowDateError), http.StatusBadRequest)
		return
	}

	// check repeat
	if repeat == "" {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidRepeatError), http.StatusBadRequest)
		return
	} else if !repeatRulePattern.MatchString(repeat) {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidRepeatError), http.StatusBadRequest)
		return
	}

	log.Println("Before nextDay")
	nextDay, err := dateutil.NextDate(dtNow, dtParsed, repeat)

	if err != nil {
		err := fmt.Errorf("wrong repeat value")
		w.WriteHeader(http.StatusInternalServerError)
		WriteResponse(w, []byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	WriteResponse(w, []byte(nextDay)) // w.Write([]byte(nextDay))
}

// WriteResponse: если происходит ошибка в методе Write, то она фатальная (panic)
func WriteResponse(w http.ResponseWriter, s []byte) {
	_, err := w.Write(s)
	if err != nil {
		log.Printf("Can not write response: %s", err.Error())
	}
}

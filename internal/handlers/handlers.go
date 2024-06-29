package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/wisdomdevil/go_final_project/internal/config"
	"github.com/wisdomdevil/go_final_project/internal/dateutil"
	"github.com/wisdomdevil/go_final_project/internal/db/repo"
	"github.com/wisdomdevil/go_final_project/internal/models"
)

const (
	MarshallingError    = "Error in marshalling JSON."
	UnMarshallingError  = "Error in unmarshalling JSON."
	ResponseWriteError  = "Error in writing data."
	ReadingError        = "Error in reading data."
	InvalidIdError      = "Invalid id."
	IdMissingError      = "ID is missing."
	InvalidDateError    = "Invalid date."
	InvalidNowDateError = "Invalid now date."
	InvalidRepeatError  = "Invalid repeat value."
	InternalServerError = "Internal server error."
	ValidatingDateError = "Error in validating date."
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
		err := fmt.Errorf("Wrong date: %v\n", err)
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
		err := fmt.Errorf("Wrong repeat value")
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
		log.Fatalf("Can not write response: %s", err.Error())
	}
}

// это совокупность хэндлеров, часто называется api
type Api struct {
	repo   *repo.TasksRepository
	config *config.Config
}

// это конструктор объекта api.
func NewApi(repo *repo.TasksRepository, config *config.Config) *Api {
	return &Api{repo: repo, config: config} // создаем ссылку на объект api со свойством repo, равным repo из параметров функции
}

func (a *Api) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		idToSearch := r.URL.Query().Get("id") // это параметр запроса

		// if idToSearch != "" {
		// не нужна проверка на пустой id, потому что strconv.Atoi в таком случае всё равно вернёт ошибку
		id, err := strconv.Atoi(idToSearch)
		if err != nil {
			log.Println("error:", err)
			RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
			return // иначе пойдем в a.GetTask(w, r, id) это стиль с гардами (защитниками). иначе надо написать else {a.GetTask(w, r, id)}
		}
		a.GetTask(w, r, id)

	case r.Method == http.MethodPost:
		log.Println("We are in MethodPost")
		a.CreateTask(w, r)
	case r.Method == http.MethodPut:
		a.UpdateTask(w, r)
	case r.Method == http.MethodDelete:
		idToSearch := r.URL.Query().Get("id") // это параметр запроса
		if idToSearch != "" {
			a.DeleteTask(w, r)
		} else {
			RenderApiErrorAndResponse(w, fmt.Errorf(IdMissingError), http.StatusBadRequest)
			return
		}
	}
}

// http://localhost:7540/api/tasks/257
func (a *Api) GetTaskByIdHandler(w http.ResponseWriter, r *http.Request) {
	idToSearch := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(IdMissingError), http.StatusBadRequest)
		return
	}
	a.GetTask(w, r, id)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("search") != "" {
		s := r.URL.Query().Get("search")
		a.SearchTasks(w, r, s)
	} else {
		a.GetAllTasks(w)
	}
}

func (a *Api) GetAllTasks(w http.ResponseWriter) {
	foundTasks, err := a.repo.GetAllTasks()
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError) // 500
		return
	}

	result := make(map[string][]models.Task)
	result["tasks"] = foundTasks

	resp, err := json.Marshal(result)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		return
	}
}

func (a *Api) SearchTasks(w http.ResponseWriter, r *http.Request, search string) {
	foundTasks, err := a.repo.SearchTasks(repo.QueryDataFromString(search))
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError) // 500
		return
	}

	result := make(map[string][]models.Task)
	result["tasks"] = foundTasks

	resp, err := json.Marshal(result)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		return
	}
}

// CreateTask posts task into DB
func (a *Api) CreateTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(ReadingError), http.StatusBadRequest) // 400
		return
	}
	log.Println("received:", buf.String())

	parseBody := models.Task{}
	err = json.Unmarshal(buf.Bytes(), &parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	err = parseBody.ValidateAndNormalizeDate()
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(ValidatingDateError), http.StatusBadRequest)
		return
	}

	id, err := a.repo.AddTask(parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	WriteResponse(w, []byte(fmt.Sprintf("{\"id\":%d}", id))) //
}

// UpdateTask updates task in DB
func (a *Api) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(ReadingError), http.StatusBadRequest) // 400
		return
	}

	parseBody := models.Task{}
	err = json.Unmarshal(buf.Bytes(), &parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	err = parseBody.ValidateAndNormalizeDate()
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(ValidatingDateError), http.StatusBadRequest)
		return
	}
	idToSearch, err := strconv.Atoi(parseBody.ID)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	_, err = a.repo.GetTask(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	err = a.repo.UpdateTaskInBd(parseBody)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	jsonItem, err := json.Marshal(parseBody)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	WriteResponse(w, jsonItem)
}

func (a *Api) DeleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("We are in DeleteTask")
	idToSearch := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	err = a.repo.DeleteTask(id)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusInternalServerError) // 500
		return
	}
	w.WriteHeader(http.StatusOK)
	WriteResponse(w, []byte("{}")) // пустой JSON
}

func (a *Api) TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("We are in TaskDoneHandler")
	idToSearch := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	newTask, err := a.repo.PostTaskDone(id)
	if newTask == nil {
		w.WriteHeader(http.StatusOK)
		WriteResponse(w, []byte("{}")) // строка с пустым json
		return
	}
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError) // 500
		return
	}
	w.WriteHeader(http.StatusOK)
	WriteResponse(w, []byte("{}")) // w.Write(resp)
}

func (a *Api) GetTask(w http.ResponseWriter, r *http.Request, id int) {
	foundTask, err := a.repo.GetTask(id)
	log.Println("we are in GetTask", "foundTask:", foundTask)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError) // 500
		return
	}

	resp, err := json.Marshal(foundTask)
	if err != nil {
		log.Println("error:", err)
		RenderApiErrorAndResponse(w, fmt.Errorf(MarshallingError), http.StatusInternalServerError) // 500
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		return
	}
}

// SigninHandler проверяет пароль и генерирует jwt token, если пароль верный
func (a *Api) SigninHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		RenderApiErrorAndResponse(w, fmt.Errorf(ReadingError), http.StatusBadRequest)
		return
	}

	// берем пароль из request Body и записываем его в структуру signinRequest{} в поле Password
	reqBody := signinRequest{}
	err = json.Unmarshal(buf.Bytes(), &reqBody)
	if err != nil {
		RenderApiErrorAndResponse(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	secret := []byte(a.config.EncryptionSecretKey)
	hashedUserPassword := HashPassword([]byte(reqBody.Password), secret)
	hashedEnvPassword := HashPassword([]byte(a.config.AppPassword), secret)

	if hashedUserPassword != hashedEnvPassword {
		RenderApiErrorAndResponse(w, fmt.Errorf("Wrong password."), http.StatusUnauthorized)
		return
	}

	// получаем подписанный токен
	tokenValue, err := createToken(reqBody.Password, a.config.EncryptionSecretKey)
	if err != nil {
		RenderApiErrorAndResponse(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
	}

	// записываем в response Body токен
	response := signinResponse{Token: tokenValue}
	respBody, err := json.Marshal(response)
	if err != nil {
		RenderApiErrorAndResponse(w, fmt.Errorf(MarshallingError), http.StatusInternalServerError)
		return
	}

	WriteResponse(w, respBody)
	w.WriteHeader(http.StatusOK)
}

// добавим проверку аутентификации для API-запросов
func (a *Api) Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r.Cookie("token")) // token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjdhMzVhZTIwYzlhZmYyY2I5MWVhYWQxODc0NzIyZDFhYzgxMzQ4MjM4OThmOTE5YWY1ZjM4ZDg4MzQ1YzFmN2YifQ.ALEVDX8JerY2jnHN1uJG8URVpsxjSr0MvNGI5P1u4ts <nil>

		// смотрим наличие пароля
		pass := a.config.AppPassword
		if len(pass) > 0 {
			var jwtFromRequest string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err != nil {
				RenderApiErrorAndResponse(w, fmt.Errorf("Empty token."), http.StatusUnauthorized)
				return
			}
			jwtFromRequest = cookie.Value

			secret := []byte(a.config.EncryptionSecretKey)

			// валидация и проверка JWT-токена
			// парсим токен
			jwtToken, err := jwt.Parse(jwtFromRequest, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				RenderApiErrorAndResponse(w, fmt.Errorf("Invalid token."), http.StatusUnauthorized)
				return
			}

			// приводим поле Claims к типу jwt.MapClaims
			res, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				RenderApiErrorAndResponse(w, fmt.Errorf("Failed to typecast to jwt.MapCalims."), http.StatusUnauthorized)
				return
			}

			// Так как jwt.Claims — словарь вида map[string]inteface{}, используем синтакис получения
			// занчения по ключу. Получаем значение ключа "password"
			pass := res["password"]
			// loginRaw — интерфейс, так как тип значения в jwt.Claims — интерфейс.
			// Чтобы получить строку, нужно снова сделать приведение типа к строке.
			_, ok = pass.(string)
			if !ok {
				RenderApiErrorAndResponse(w, fmt.Errorf("Failed to typecast to string."), http.StatusUnauthorized)
				return
			}
			// fmt.Println(password)
		}
		next(w, r)
	})
}

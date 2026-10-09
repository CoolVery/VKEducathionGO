package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFindUsers(t *testing.T) {
	t.Run("Валидация параметров", func(t *testing.T) {
		t.Run("Limit меньше 0", testFindUser_LimitValidationSmall0)
		t.Run("Offset меньше 0", testFindUser_OffsetValidationSmall0)
	})

	t.Run("Успешные ответы", func(t *testing.T) {
		t.Run("Пустой список пользователей", testFindUser_CheckNilUsers)
		t.Run("Limit и NextPage", testFindUser_CheckLimitAndNextPage)
	})

	t.Run("Ошибки HTTP", func(t *testing.T) {
		t.Run("Timeout", testFindUser_CheckErrTimeOut)
		t.Run("Unknown error", testFindUser_CheckErrUnknown)
		t.Run("Unauthorized", testFindUser_CheckStatusUnauthorized)
		t.Run("Internal Server Error", testFindUser_CheckStatusInternalServerError)
	})

	t.Run("Ошибки парсинга и Bad Request", func(t *testing.T) {
		t.Run("BadRequest с nil телом", testFindUser_CheckStatusStatusBadRequestNil)
		t.Run("BadRequest ErrorBadOrderField", testFindUser_CheckStatusStatusBadRequestErrorBadOrderField)
		t.Run("BadRequest неизвестная ошибка", testFindUser_CheckStatusStatusBadRequestUnrnown)
		t.Run("Не удалось распарсить result", testFindUser_CheckFailUnpackResult)
	})
}

func testFindUser_LimitValidationSmall0(t *testing.T) {
	srv := &SearchClient{}

	_, err := srv.FindUsers(SearchRequest{Limit: -1})
	if err == nil {
		t.Fatal("Ожидали ошибку, пришел nil")
	}
	if err.Error() != "limit must be > 0" {
		t.Fatalf("Ожидали ошибку: limit must be > 0, но пришло: %s", err.Error())
	}
}
func testFindUser_OffsetValidationSmall0(t *testing.T) {
	srv := &SearchClient{}

	_, err := srv.FindUsers(SearchRequest{Offset: -1})
	if srv == nil {
		t.Fatal("Ожидали ошибку, пришел nil")
	}
	if err.Error() != "offset must be > 0" {
		t.Fatalf("Ожидали ошибку: offset must be > 0, но пришло: %s", err.Error())
	}
}
func testFindUser_CheckNilUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(users.Users) != 0 {
		t.Fatalf("Ожидали длину 0, а пришло: %d", len(users.Users))
	}
	if users.NextPage {
		t.Fatalf("Ожидали, что следующая страница будет false")
	}
}
func testFindUser_CheckErrTimeOut(t *testing.T) {
	old := client
	client = &http.Client{Timeout: time.Millisecond * 50}
	defer func() { client = old }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 200)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку timeout")
	}
	if !strings.Contains(err.Error(), "timeout for") {
		t.Fatalf("Ожидали ошибку timeiout, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckErrUnknown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку unknown")
	}
	if !strings.Contains(err.Error(), "unknown error") {
		t.Fatalf("Ожидали ошибку unknown, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckStatusUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку bad AccessToken")
	}
	if !strings.Contains(err.Error(), "bad AccessToken") {
		t.Fatalf("Ожидали ошибку bad AccessToken, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку SearchServer fatal error")
	}
	if !strings.Contains(err.Error(), "SearchServer fatal error") {
		t.Fatalf("Ожидали ошибку SearchServer fatal error, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckStatusStatusBadRequestNil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку cant unpack error json")
	}
	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Fatalf("Ожидали ошибку cant unpack error json, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckStatusStatusBadRequestErrorBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: ErrorBadOrderField})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку OrderFeld")
	}
	if !strings.Contains(err.Error(), "OrderFeld") {
		t.Fatalf("Ожидали ошибку OrderFeld, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckStatusStatusBadRequestUnrnown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "bad string"})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку unknown bad request error")
	}
	if !strings.Contains(err.Error(), "unknown bad request error") {
		t.Fatalf("Ожидали ошибку unknown bad request error, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckFailUnpackResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("fdfdfd")
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 1, Offset: 2}
	users, err := srv.FindUsers(req)
	if err == nil {
		t.Fatal("Ожидали ошибку cant unpack result json")
	}
	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Fatalf("Ожидали ошибку cant unpack result json, а пришло: %s", err.Error())
	}
	if users != nil {
		t.Fatalf("Ожидали, users будет nil, а пришло: %v", users)
	}
}
func testFindUser_CheckLimitAndNextPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]User{
			{ID: 1, Name: "Иван", Age: 30, About: "любит Go", Gender: "m"},
			{ID: 2, Name: "Мария", Age: 25, About: "любит тесты", Gender: "f"},
			{ID: 3, Name: "Пётр", Age: 41, About: "пишет на C++", Gender: "m"},
			{ID: 4, Name: "Анна", Age: 22, About: "учится в вузе", Gender: "f"},
			{ID: 5, Name: "Сергей", Age: 35, About: "backend-разработчик", Gender: "m"},
			{ID: 6, Name: "Ольга", Age: 28, About: "дизайнер интерфейсов", Gender: "f"},
			{ID: 7, Name: "Дмитрий", Age: 33, About: "DevOps-инженер", Gender: "m"},
			{ID: 8, Name: "Екатерина", Age: 27, About: "аналитик данных", Gender: "f"},
			{ID: 9, Name: "Алексей", Age: 45, About: "тимлид", Gender: "m"},
			{ID: 10, Name: "Наталья", Age: 31, About: "QA-инженер", Gender: "f"},
			{ID: 11, Name: "Николай", Age: 38, About: "архитектор", Gender: "m"},
			{ID: 12, Name: "Татьяна", Age: 29, About: "product-менеджер", Gender: "f"},
			{ID: 13, Name: "Андрей", Age: 24, About: "мобильный разработчик", Gender: "m"},
			{ID: 14, Name: "Елена", Age: 36, About: "HR-специалист", Gender: "f"},
			{ID: 15, Name: "Михаил", Age: 42, About: "системный администратор", Gender: "m"},
			{ID: 16, Name: "Ирина", Age: 26, About: "frontend-разработчик", Gender: "f"},
			{ID: 17, Name: "Владимир", Age: 50, About: "преподаватель", Gender: "m"},
			{ID: 18, Name: "Светлана", Age: 34, About: "маркетолог", Gender: "f"},
			{ID: 19, Name: "Роман", Age: 23, About: "стажёр-программист", Gender: "m"},
			{ID: 20, Name: "Юлия", Age: 39, About: "бухгалтер", Gender: "f"},
			{ID: 21, Name: "Артём", Age: 32, About: "data-инженер", Gender: "m"},
			{ID: 22, Name: "Ксения", Age: 21, About: "студентка", Gender: "f"},
			{ID: 23, Name: "Максим", Age: 44, About: "руководитель отдела", Gender: "m"},
			{ID: 24, Name: "Вероника", Age: 30, About: "технический писатель", Gender: "f"},
			{ID: 25, Name: "Игорь", Age: 37, About: "сетевой инженер", Gender: "m"},
			{ID: 26, Name: "Алиса", Age: 28, About: "специалист по безопасности", Gender: "f"},
		})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL:         ts.URL,
		AccessToken: "test-token",
	}
	req := SearchRequest{Limit: 100, Offset: 2}
	res, err := srv.FindUsers(req)
	if res.NextPage == false {
		t.Fatal("Ожидали true в NextPage")
	}
	if len(res.Users) != 25 {
		t.Fatalf("Ожидали возвращения списка с 25 элементов, а пришло: %d --- %v", len(res.Users), res.Users)
	}
	if err != nil {
		t.Fatal("Ожидали nil в err")
	}
}

func TestSearchServer(t *testing.T) {
	t.Run("Параметр invalid", func(t *testing.T) {
		t.Run("Offset invalid", testSearchServer_OffsetInvalid)
		t.Run("Limit invalid", testSearchServer_LimitInvalid)
		t.Run("OrderField invalid", testSearchServer_OrderFieldInvalid)
		t.Run("OrderBy invalid not int", testSearchServer_OrderByInvalidNotInt)
		t.Run("OrderBy invalid not access", testSearchServer_OrderByInvalidNotAccess)
	})
	t.Run("Чтение из файла", func(t *testing.T) {
		t.Run("1 пользователь со всеми полями", testSearchServer_QueryOneUserWithAllField)
		t.Run("34 пользователя с пустым query", testSearchServer_QueryAllUsersWithNilName)

	})
}

func testSearchServer_OffsetInvalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?offset=ddd")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Ожидали код 400, а пришло: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Offset invalid") {
		t.Fatalf("Ожидали в теле ошибку Offset invalid, а получили: %s", body)
	}
}
func testSearchServer_LimitInvalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?limit=ddd")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Ожидали код 400, а пришло: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Limit invalid") {
		t.Fatalf("Ожидали в теле ошибку Limit invalid, а получили: %s", body)
	}
}
func testSearchServer_OrderFieldInvalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?order_field=About")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Ожидали код 400, а пришло: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "OrderField invalid") {
		t.Fatalf("Ожидали в теле ошибку OrderField invalid, а получили: %s", body)
	}
}
func testSearchServer_OrderByInvalidNotInt(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?order_by=AAA")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Ожидали код 400, а пришло: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "OrderBy invalid") {
		t.Fatalf("Ожидали в теле ошибку OrderBy invalid, а получили: %s", body)
	}
}
func testSearchServer_OrderByInvalidNotAccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?order_by=3")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Ожидали код 400, а пришло: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "OrderBy invalid") {
		t.Fatalf("Ожидали в теле ошибку OrderBy invalid, а получили: %s", body)
	}
}
func testSearchServer_QueryOneUserWithAllField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?order_by=-1&order_field=age&limit=1&offset=0&query=on")
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()
	var users []User
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		t.Fatalf("Ожидали список пользователей, а пришло: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Ожидали, что вернется 1 пользователь, а пришло: %d", len(users))
	}
}
func testSearchServer_QueryAllUsersWithNilName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Запрос вообще не прошел: %v", err)
	}
	defer resp.Body.Close()
	var users []User
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		t.Fatalf("Ожидали список пользователей, а пришло: %v", err)
	}
	if len(users) != 35 {
		t.Fatalf("Ожидали, что вернется 35 пользователя, а пришло: %d", len(users))
	}
}
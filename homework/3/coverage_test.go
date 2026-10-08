package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)



type TestCase struct {
	ID      string
	Result  []User
	IsError bool
}

func TestFindUser_LimitValidationSmall0(t *testing.T) {
	srv := &SearchClient{}

	_, err := srv.FindUsers(SearchRequest{Limit: -1})
	if err == nil {
		t.Fatal("Ожидали ошибку, пришел nil")
	}
	if err.Error() != "limit must be > 0" {
		t.Fatalf("Ожидали ошибку: limit must be > 0, но пришло: %s", err.Error())
	}
}
func TestFindUser_OffsetValidationSmall0(t *testing.T) {
	srv := &SearchClient{}

	_, err := srv.FindUsers(SearchRequest{Offset: -1})
	if srv == nil {
		t.Fatal("Ожидали ошибку, пришел nil")
	}
	if err.Error() != "offset must be > 0" {
		t.Fatalf("Ожидали ошибку: offset must be > 0, но пришло: %s", err.Error())
	}
}
func TestFindUser_CheckNilUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckErrTimeOut(t *testing.T) {
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
		URL: ts.URL,
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
func TestFindUser_CheckErrUnknown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckStatusUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode([]User{})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckStatusStatusBadRequestNil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckStatusStatusBadRequestErrorBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: ErrorBadOrderField})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckStatusStatusBadRequestUnrnown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "bad string"})
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckFailUnpackResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("fdfdfd")
	}))
	defer ts.Close()

	srv := &SearchClient{
		URL: ts.URL,
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
func TestFindUser_CheckLimitAndNextPage(t *testing.T) {
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
		URL: ts.URL,
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
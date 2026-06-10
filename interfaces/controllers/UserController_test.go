package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/dto"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createAddUserRequest(method string, target string, body io.Reader, contentType string) *http.Request {
	request := httptest.NewRequest(
		method,
		target,
		body,
	)
	request.Header.Set("Content-Type", contentType)
	return request
}

func Test_addUser(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}

	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "Add valid user",
			args: args{
				w: httptest.NewRecorder(),
				r: createAddUserRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email@gmail.com"
					}`),
					"application/json",
				),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Add user with invalid content-type",
			args: args{
				w: httptest.NewRecorder(),
				r: createAddUserRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email@gmail.com"
					}`),
					"",
				),
			},
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "Add invalid user",
			args: args{
				w: httptest.NewRecorder(),
				r: createAddUserRequest(
					http.MethodPost,
					"/user/add",
					bytes.NewBufferString(`{
						"Username":  "Edgouille",
						"Password":  "MDP!@#111111111",
						"Email":     "email"
					}`),
					"application/json",
				),
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				services.UserRepository = new(memory.InMemoryUserRepository)
			})

			addUser(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatus {
				t.Errorf("addUser() got status code = %v, want = %v",
					tt.args.w.Code, tt.wantStatus)
			}
		})
	}
}

func createGetUserRequest(
	method string, target string, body io.Reader,
	pathParam string, pathParamValue string,
) *http.Request {
	request := httptest.NewRequest(method, target, body)

	if pathParam != "" || pathParamValue != "" {
		request.SetPathValue(pathParam, pathParamValue)
	}

	return request
}

func Test_getUser(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}

	creationTime := time.Now()

	userToAdd := dto.UserDTO{
		Username: "Edgouille",
		Password: "MDP!@#111111111",
		Email:    "email@email.com",
	}
	e := services.AddUser(userToAdd)
	if e != nil {
		t.Fatal(e.Error())
	}

	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "Get user",
			args: args{
				w: httptest.NewRecorder(),
				r: createGetUserRequest(http.MethodGet, "/user/", nil, "username", "Edgouille"),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       fmt.Sprintf(`{"username":"Edgouille","email":"email@email.com","created_at":"%v","updated_at":"%v"}`+"\n", creationTime.Format(time.RFC3339Nano), creationTime.Format(time.RFC3339Nano)),
		},
		{
			name: "Get nonexistent user",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/user/10001", nil),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "user not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				services.UserRepository = new(memory.InMemoryUserRepository)
			})

			getUser(tt.args.w, tt.args.r)
			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getUser() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getUser() got body = %v, want = %v", tt.args.w.Body, tt.wantBody)
			}
		})
	}
}

package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ValidateExpectedCodeAndBody(w *httptest.ResponseRecorder, t *testing.T, expectedCode int, expectedBody string) {
	if w.Code != expectedCode {
		t.Errorf("Expected %d, got %d", expectedCode, w.Code)
	}

	if w.Body.String() != expectedBody {
		t.Errorf("Expected %s, got %s", expectedBody, w.Body.String())
	}
}

func CreateRequest(method string, target string, body io.Reader, contentType string) *http.Request {
	request := httptest.NewRequest(
		method,
		target,
		body,
	)
	request.Header.Set("Content-Type", contentType)
	return request
}

func CreateRequestWithPathParam(
	method string, target string, body io.Reader,
	pathParam string, pathParamValue string,
) *http.Request {
	request := httptest.NewRequest(method, target, body)

	request.SetPathValue(pathParam, pathParamValue)

	return request
}

func CreateTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func CreateTimedOutTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	t.Cleanup(cancel)
	return ctx
}

func CreateCancelledTestContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

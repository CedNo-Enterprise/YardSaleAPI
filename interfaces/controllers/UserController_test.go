package controllers

import (
	"net/http"
	"testing"
)

func TestAddUserHandlersToMux(t *testing.T) {
	type args struct {
		mux *http.ServeMux
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddUserHandlersToMux(tt.args.mux)
		})
	}
}

func Test_addUser(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addUser(tt.args.w, tt.args.r)
		})
	}
}

func Test_getUser(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUser(tt.args.w, tt.args.r)
		})
	}
}

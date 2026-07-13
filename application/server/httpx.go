package server

import (
	"GarageSaleAPI/application/server/apperror"
	"errors"
	"net/http"
)

var statusByKind = map[apperror.Kind]int{
	apperror.KindNotFound:  http.StatusNotFound,
	apperror.KindInvalid:   http.StatusBadRequest,
	apperror.KindConflict:  http.StatusConflict,
	apperror.KindForbidden: http.StatusForbidden,
	apperror.KindInternal:  http.StatusInternalServerError,
}

func WriteError(w http.ResponseWriter, err error) {
	if appErr, ok1 := errors.AsType[*apperror.AppError](err); ok1 {
		status, ok2 := statusByKind[appErr.Kind]
		if !ok2 {
			status = http.StatusInternalServerError
		}
		http.Error(w, appErr.Message, status)
		return
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

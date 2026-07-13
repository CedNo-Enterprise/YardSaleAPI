package apperror

type Kind string

const (
	KindNotFound  Kind = "not_found"
	KindInvalid   Kind = "invalid"
	KindConflict  Kind = "conflict"
	KindForbidden Kind = "forbidden"
	KindInternal  Kind = "internal"
)

type AppError struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(msg string, err error) *AppError {
	return &AppError{Kind: KindNotFound, Message: msg, Err: err}
}

func Invalid(msg string, err error) *AppError {
	return &AppError{Kind: KindInvalid, Message: msg, Err: err}
}

func Conflict(msg string, err error) *AppError {
	return &AppError{Kind: KindConflict, Message: msg, Err: err}
}

func Forbidden(msg string, err error) *AppError {
	return &AppError{Kind: KindForbidden, Message: msg, Err: err}
}

func Internal(err error) *AppError {
	return &AppError{Kind: KindInternal, Message: "internal error", Err: err}
}

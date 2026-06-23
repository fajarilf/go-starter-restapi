package domain

type Kind int

const (
	KindValidation = iota
	KindNotFound
	KindConflict
	KindInternal
)

type AppError struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NewValidationError(msg string) *AppError {
	return &AppError{Kind: KindValidation, Message: msg}
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{Kind: KindNotFound, Message: msg}
}

func NewInternalError(msg string) *AppError {
	return &AppError{Kind: KindInternal, Message: msg}
}

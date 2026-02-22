package domain

// ErrValidation is returned when a business rule or validation fails (Maps to HTTP 400 Bad Request)
type ErrValidation struct {
	Message string
}

func (e *ErrValidation) Error() string {
	return e.Message
}

// ErrNotFound is returned when a requested resource does not exist (Maps to HTTP 404 Not Found)
type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return e.Message
}

// ErrConflict is returned when an action violates the current state, like duplicates (Maps to HTTP 409 Conflict)
type ErrConflict struct {
	Message string
}

func (e *ErrConflict) Error() string {
	return e.Message
}

// ErrUnauthorized is returned when authentication or authorization fails (Maps to HTTP 401 Unauthorized)
type ErrUnauthorized struct {
	Message string
}

func (e *ErrUnauthorized) Error() string {
	return e.Message
}

// ErrInternal is returned for unexpected system-level faults (Maps to HTTP 500 Internal Server Error)
type ErrInternal struct {
	Message string
}

func (e *ErrInternal) Error() string {
	return e.Message
}

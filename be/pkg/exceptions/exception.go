package exceptions

import "net/http"

// Stable error codes. The frontend translates them via `errors.<CODE>` keys,
// so never change an existing code, only add new ones.
const (
	CodeBadRequest             = "BAD_REQUEST"
	CodeValidationFailed       = "VALIDATION_FAILED"
	CodeUnauthorized           = "AUTH_UNAUTHORIZED"
	CodeTokenExpired           = "AUTH_TOKEN_EXPIRED"
	CodeGoogleExchangeFailed   = "AUTH_GOOGLE_EXCHANGE_FAILED"
	CodeGoogleInvalidToken     = "AUTH_GOOGLE_INVALID_TOKEN"
	CodeGoogleEmailUnverified  = "AUTH_GOOGLE_EMAIL_UNVERIFIED"
	CodeGoogleNotConfigured    = "AUTH_GOOGLE_NOT_CONFIGURED"
	CodeRefreshInvalid         = "AUTH_REFRESH_INVALID"
	CodeRefreshReused          = "AUTH_REFRESH_REUSED"
	CodeForbidden              = "FORBIDDEN"
	CodeNotFound               = "NOT_FOUND"
	CodeRateLimited            = "RATE_LIMITED"
	CodeInternal               = "INTERNAL_SERVER_ERROR"
	CodeUserInvalidTimezone    = "USER_INVALID_TIMEZONE"
	CodeServiceUnavailable     = "SERVICE_UNAVAILABLE"
	CodeUserInvalidSteps       = "USER_INVALID_STEPS"
	CodeDeckNotFound           = "DECK_NOT_FOUND"
	CodeDeckNameTaken          = "DECK_NAME_TAKEN"
	CodeCardNotFound           = "CARD_NOT_FOUND"
	CodeCardDuplicate          = "CARD_DUPLICATE_HEADWORD"
	CodeStudyStateChanged      = "STUDY_CARD_STATE_CHANGED"
	CodeStudyNothingToUndo     = "STUDY_NOTHING_TO_UNDO"
	CodeVocabNoEntries         = "VOCAB_NO_ENTRIES"
	CodeVocabTooManyEntries    = "VOCAB_TOO_MANY_ENTRIES"
	CodeVocabQuotaExceeded     = "VOCAB_QUOTA_EXCEEDED"
	CodeVocabJobNotFound       = "VOCAB_JOB_NOT_FOUND"
	CodeVocabItemNotFound      = "VOCAB_ITEM_NOT_FOUND"
	CodeVocabItemNotReady      = "VOCAB_ITEM_NOT_READY"
	CodeVocabNothingToSave     = "VOCAB_NOTHING_TO_SAVE"
	CodeVocabNotConfigured     = "VOCAB_GENERATOR_NOT_CONFIGURED"
	CodeVocabGenerationFailed  = "VOCAB_GENERATION_FAILED"
	CodeVocabImageSearchFailed = "VOCAB_IMAGE_SEARCH_FAILED"
	CodeLexemeNotFound         = "LEXEME_NOT_FOUND"
	CodeSuggestionLimit        = "VOCAB_SUGGESTION_LIMIT"
	CodeSuggestionFailed       = "VOCAB_SUGGESTION_FAILED"
	CodeSuggestionNotFound     = "VOCAB_SUGGESTION_NOT_FOUND"
	CodeTopicRequired          = "VOCAB_TOPIC_REQUIRED"
	CodeImageNotFound          = "IMAGE_NOT_FOUND"
)

var (
	ErrBadRequest = func(message string, err error) *AppError {
		return NewException(http.StatusBadRequest, CodeBadRequest, message, err)
	}
	ErrUnauthorized = func(message string, err error) *AppError {
		return NewException(http.StatusUnauthorized, CodeUnauthorized, message, err)
	}
	ErrInternalServerError = func(message string, err error) *AppError {
		return NewException(http.StatusInternalServerError, CodeInternal, message, err)
	}
	ErrNotFound = func(message string, err error) *AppError {
		return NewException(http.StatusNotFound, CodeNotFound, message, err)
	}
	ErrForbidden = func(message string, err error) *AppError {
		return NewException(http.StatusForbidden, CodeForbidden, message, err)
	}
)

type AppError struct {
	Message string            `json:"message"`
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Params  map[string]any    `json:"params,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
	Err     error             `json:"-"`
}

func NewException(status int, code string, message string, err error) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithCode returns a copy of the error with a more specific error code.
func (e *AppError) WithCode(code string) *AppError {
	cp := *e
	cp.Code = code
	return &cp
}

func (e *AppError) WithParams(params map[string]any) *AppError {
	cp := *e
	cp.Params = params
	return &cp
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

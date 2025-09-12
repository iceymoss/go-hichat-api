package xerr

const (
	SERVER_COMMON_ERROR = 100001
	REQUEST_PARAM_ERROR = 100002
	TOKEN_EXPIRE_ERROR  = 100003
	DB_ERROR            = 100004

	ErrInternalServer = 500 // HTTP 500

	ErrBadRequest       = 1000 // HTTP 400
	ErrInvalidInput     = 1001 // HTTP 400
	ErrMissingParameter = 1002 // HTTP 400
	ErrInvalidJSON      = 1003 // HTTP 400

	ErrUnauthenticated  = 1100 // HTTP 401
	ErrInvalidToken     = 1101 // HTTP 401
	ErrTokenExpired     = 1102 // HTTP 401
	ErrInvalidSignature = 1103 // HTTP 401

	ErrForbidden        = 1200 // HTTP 403
	ErrInsufficientPriv = 1201 // HTTP 403
	ErrReadOnlyMode     = 1202 // HTTP 403

	ErrNotFound         = 1300 // HTTP 404
	ErrResourceNotFound = 1301 // HTTP 404
	ErrEndpointRemoved  = 1302 // HTTP 410
)

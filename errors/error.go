package lanky_errors

import (
	"fmt"
	"net/http"
)

// LankyErrorCode represents an error code specific to the Lanky library.
type LankyErrorCode uint64

// LankyCommonError represents a common error structure used in the Lanky library.
type LankyCommonError struct {
	ClientMessage string         `json:"message"`
	SystemMessage any            `json:"data"`
	Code          LankyErrorCode `json:"code"`
	Err           *string        `json:"-"`
	Trace         *string        `json:"-"`
}

// Error returns a string representation of the LankyCommonError.
// It formats the error message and the error trace.
// The error message is obtained from the Err field of the LankyCommonError struct.
// The error trace is obtained from the Trace field of the LankyCommonError struct.
// The formatted string includes the error message and the error trace.
// The error message is prefixed with "Err: " and the error trace is prefixed with "Trace: ".
// The formatted string is returned as the error message.
func (lce *LankyCommonError) Error() string {
	return fmt.Sprintf("Err: %+v\nTrace: %+v", lce.Err, lce.Trace)
}

// SetClientMessage sets the client message for the LankyCommonError.
// It takes a string value as input and assigns it to the ClientMessage field of the LankyCommonError struct.
func (lce *LankyCommonError) SetClientMessage(val string) {
	lce.ClientMessage = val
}

// SetSystemMessage sets the system message of the LankyCommonError.
// The system message is a value that provides additional information about the error.
// It can be of any type.
func (lce *LankyCommonError) SetSystemMessage(val any) {
	lce.SystemMessage = val
}

// LankyHttpCommonError represents a common error structure for HTTP-related errors in the Lanky library.
type LankyHttpCommonError struct {
	LankyCommonError     // Embedded struct for common error fields
	HttpStatusNumber int `json:"-"` // HTTP status code associated with the error
}

// Error returns a string representation of the LankyHttpCommonError.
// It formats the error message and the HTTP trace information.
// The returned string includes the error message and the HTTP trace.
func (lce *LankyHttpCommonError) Error() string {
	return fmt.Sprintf("HttpErr: %+v\nHttpTrace: %+v", lce.Err, lce.Trace)
}

// UnidentifiedError represents an unidentified error in the Lanky library.
const UnidentifiedError LankyErrorCode = 0

// mapError represents a collection of Lanky error codes and their corresponding common errors.
type mapError struct {
	dict map[LankyErrorCode]*LankyCommonError
	stat map[LankyErrorCode]int
}

// me is a global variable that represents a mapError instance.
// It is used to store LankyErrorCode keys and corresponding LankyCommonError values.
// The `dict` field is a map that maps LankyErrorCode keys to LankyCommonError values.
// The `stat` field is a map that stores the count of occurrences of each LankyErrorCode.
var me = &mapError{
	dict: make(map[LankyErrorCode]*LankyCommonError),
	stat: make(map[LankyErrorCode]int),
}

// Register registers the given dictionary of Lanky error codes and their corresponding LankyCommonError objects,
// as well as the statistics map of Lanky error codes and their corresponding HTTP status codes.
// It initializes the mapError instance with the provided dictionary and statistics.
// Additionally, it adds an entry for the UnidentifiedError code in both the dictionary and statistics map,
// with the corresponding LankyCommonError object and HTTP status code for internal server error.
func Register(dict map[LankyErrorCode]*LankyCommonError, stat map[LankyErrorCode]int) {
	me = &mapError{
		dict: dict,
		stat: stat,
	}

	me.dict[UnidentifiedError] = &LankyCommonError{
		ClientMessage: "Unidentified error has occured. Please contact our dev",
		SystemMessage: "Internal server error",
		Code:          UnidentifiedError,
	}

	me.stat[UnidentifiedError] = http.StatusInternalServerError
}

// GetHttpStatus returns the HTTP status code associated with the LankyCommonError.
// If the status code is not found in the internal map, it returns http.StatusInternalServerError.
func (lce *LankyCommonError) GetHttpStatus() int {
	if st := me.stat[lce.Code]; st == 0 {
		return http.StatusInternalServerError
	} else {
		return st
	}
}

// ToHttpStatusError converts a LankyCommonError to a LankyHttpCommonError with the corresponding HTTP status number.
// It returns a pointer to the converted LankyHttpCommonError.
func (lce *LankyCommonError) ToHttpStatusError() *LankyHttpCommonError {
	return &LankyHttpCommonError{
		LankyCommonError: *lce,
		HttpStatusNumber: lce.GetHttpStatus(),
	}
}

// New creates a new instance of LankyCommonError with the given error code and error.
// It returns a pointer to the created LankyCommonError.
// If the error is not nil, it sets the error message and error trace in the LankyCommonError.
// If the error code is UnidentifiedError, it sets the client message and system message to the error message and error trace respectively.
// If the error is already an instance of LankyCommonError, it returns the error as is.
func New(code LankyErrorCode, err error) *LankyCommonError {
	var (
		em *string
		et *string

		cm = "Unidentified error has occured. Please contact our dev"
		sm = "Internal server error"

		lce = me.dict[code]
	)

	if err != nil {
		m := err.Error()
		em = &m

		m2 := fmt.Sprintf("%+v", err)
		et = &m2

		if code == UnidentifiedError {
			cm = m
			sm = m2
		}
	}

	if lce == nil {
		lce = &LankyCommonError{
			ClientMessage: cm,
			SystemMessage: sm,
			Code:          code,
			Err:           em,
			Trace:         et,
		}
	} else {
		lce.Err = em
		lce.Trace = et
	}

	if lce2, ok := err.(*LankyCommonError); ok {
		return lce2
	}

	return lce
}

package t_api

import (
	"fmt"

	"github.com/resonatehq/resonate/pkg/promise"
	"google.golang.org/grpc/codes"
)

// ResponseStatus is the status code for the response.
type ResponseStatus int

// Application level status (2000-4999)
const (
	StatusOK        ResponseStatus = 2000
	StatusCreated   ResponseStatus = 2010
	StatusNoContent ResponseStatus = 2040

	StatusFieldValidationFailure ResponseStatus = 4000

	StatusPromiseAlreadyResolved ResponseStatus = 4030
	StatusPromiseAlreadyRejected ResponseStatus = 4031
	StatusPromiseAlreadyCanceled ResponseStatus = 4032
	StatusPromiseAlreadyTimedout ResponseStatus = 4033
	StatusLockAlreadyAcquired    ResponseStatus = 4034
	StatusTaskAlreadyClaimed     ResponseStatus = 4035
	StatusTaskAlreadyCompleted   ResponseStatus = 4036
	StatusTaskInvalidCounter     ResponseStatus = 4037
	StatusTaskInvalidState       ResponseStatus = 4038

	StatusPromiseNotFound  ResponseStatus = 4040
	StatusScheduleNotFound ResponseStatus = 4041
	StatusLockNotFound     ResponseStatus = 4042
	StatusTaskNotFound     ResponseStatus = 4043

	StatusPromiseAlreadyExists  ResponseStatus = 4090
	StatusScheduleAlreadyExists ResponseStatus = 4091
)

// String returns the string representation of the status code.
func (s ResponseStatus) String() string {
	switch s {
	case StatusOK, StatusCreated, StatusNoContent:
		return "The request was successful"
	case StatusFieldValidationFailure:
		return "The request is invalid"
	case StatusPromiseAlreadyResolved:
		return "The promise has already been resolved"
	case StatusPromiseAlreadyRejected:
		return "The promise has already been rejected"
	case StatusPromiseAlreadyCanceled:
		return "The promise has already been canceled"
	case StatusPromiseAlreadyTimedout:
		return "The promise has already timedout"
	case StatusLockAlreadyAcquired:
		return "The lock is already acquired"
	case StatusTaskAlreadyClaimed:
		return "The task is already claimed"
	case StatusTaskAlreadyCompleted:
		return "The task is already completed"
	case StatusTaskInvalidCounter:
		return "The task counter is invalid"
	case StatusTaskInvalidState:
		return "The lock state is invalid"
	case StatusPromiseNotFound:
		return "The specified promise was not found"
	case StatusScheduleNotFound:
		return "The specified schedule was not found"
	case StatusLockNotFound:
		return "The specified lock was not found"
	case StatusTaskNotFound:
		return "The specified task was not found"
	case StatusPromiseAlreadyExists:
		return "A promise with this identifier already exists"
	case StatusScheduleAlreadyExists:
		return "A schedule with this identifier already exists"
	default:
		panic(fmt.Sprintf("unknown status code %d", s))
	}
}

// HTTP maps to http status code.
func (s ResponseStatus) HTTP() int {
	return int(s) / 10
}

// GRPC maps to grpc status code.
func (s ResponseStatus) GRPC() codes.Code {
	switch s {
	case StatusOK, StatusCreated, StatusNoContent:
		return codes.OK
	case StatusFieldValidationFailure:
		return codes.InvalidArgument
	case StatusPromiseAlreadyResolved, StatusPromiseAlreadyRejected, StatusPromiseAlreadyCanceled, StatusPromiseAlreadyTimedout:
		return codes.PermissionDenied
	case StatusLockAlreadyAcquired:
		return codes.PermissionDenied
	case StatusTaskAlreadyClaimed, StatusTaskAlreadyCompleted, StatusTaskInvalidCounter, StatusTaskInvalidState:
		return codes.PermissionDenied
	case StatusPromiseAlreadyExists, StatusScheduleAlreadyExists:
		return codes.AlreadyExists
	case StatusPromiseNotFound, StatusLockNotFound, StatusTaskNotFound:
		return codes.NotFound
	default:
		panic(fmt.Sprintf("invalid status: %d", s))
	}
}

func ForbiddenStatus(state promise.State) ResponseStatus {
	switch state {
	case promise.Resolved:
		return StatusPromiseAlreadyResolved
	case promise.Rejected:
		return StatusPromiseAlreadyRejected
	case promise.Canceled:
		return StatusPromiseAlreadyCanceled
	case promise.Timedout:
		return StatusPromiseAlreadyTimedout
	default:
		panic(fmt.Sprintf("invalid promise state: %s", state))
	}
}

// Platform level errors (5000-5999)
const (
	ErrInternalServer               ResonateErrorCode = 5000
	ErrAIOStoreFailure              ResonateErrorCode = 5001
	ErrAIOStoreSerializationFailure ResonateErrorCode = 5002
	ErrSystemShuttingDown           ResonateErrorCode = 5030
	ErrAPISubmissionQueueFull       ResonateErrorCode = 5031
	ErrAIOSubmissionQueueFull       ResonateErrorCode = 5032
	ErrSchedulerQueueFull           ResonateErrorCode = 5033
)

type ResonateErrorCode int

func (e ResonateErrorCode) HTTP() int {
	return int(e) / 10
}

func (e ResonateErrorCode) GRPC() codes.Code {
	switch e {
	case ErrInternalServer:
		return codes.Internal
	case ErrAIOStoreFailure:
		return codes.Internal
	case ErrAIOStoreSerializationFailure:
		return codes.Internal
	case ErrSystemShuttingDown:
		return codes.Unavailable
	case ErrAPISubmissionQueueFull:
		return codes.Unavailable
	case ErrAIOSubmissionQueueFull:
		return codes.Unavailable
	case ErrSchedulerQueueFull:
		return codes.Unavailable
	default:
		panic(fmt.Sprintf("invalid error code: %d", e))
	}
}

type ResonateError struct {
	code          ResonateErrorCode
	reason        string
	originalError error
}

func NewResonateError(code ResonateErrorCode, out string, in error) *ResonateError {
	return &ResonateError{
		code:          code,
		reason:        out,
		originalError: in,
	}
}

func (e *ResonateError) Error() string {
	return e.reason
}

func (e *ResonateError) Unwrap() error {
	return e.originalError
}

func (e *ResonateError) Code() ResonateErrorCode {
	return e.code
}

func (e *ResonateError) Is(target error) bool {
	_, ok := target.(*ResonateError)
	return ok
}

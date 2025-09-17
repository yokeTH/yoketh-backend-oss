package apperror

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/gofiber/fiber/v3"
)

type detail struct {
	Message string      `json:"message"`
	Status  ErrorStatus `json:"status"`
}

type Response struct {
	Error detail `json:"error"`
}

type AppError struct {
	Code    int
	Message string
	Status  ErrorStatus
	Err     error
	Stack   []string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("Message: %s - Internal Error: %v", e.Message, e.Err)
}

func (e *AppError) toResponse() Response {
	return Response{
		Error: detail{
			Message: e.Message,
			Status:  e.Status,
		},
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func New(err error, code int, message string, status ErrorStatus) *AppError {
	stack := captureStack()
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Stack:   stack,
	}
}

// Adapted from https://github.com/pkg/errors/blob/5dd12d0cfe7f152f80558d591504ce685299311e/stack.go
func captureStack() []string {
	const depth = 16
	var pcs [depth]uintptr

	// Skip 4 stack frames: apperror.captureStack x2, apperror.New, apperror.InternalServerError or similar
	n := runtime.Callers(4, pcs[:])

	stackTrace := make([]string, 0, n)
	for i := range n {
		fn := runtime.FuncForPC(pcs[i])
		file, line := fn.FileLine(pcs[i])
		stackTrace = append(stackTrace, fmt.Sprintf("%s\n\tat %s:%d", fn.Name(), file, line))
	}

	return stackTrace
}

func InternalServerError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusInternalServerError, msg, status)
}

func BadRequestError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusBadRequest, msg, status)
}

func UnauthorizedError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusUnauthorized, msg, status)
}

func ForbiddenError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusForbidden, msg, status)
}

func NotFoundError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusNotFound, msg, status)
}

func ConflictError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusConflict, msg, status)
}

func UnprocessableEntityError(err error, msg string, status ErrorStatus) *AppError {
	return New(err, fiber.StatusUnprocessableEntity, msg, status)
}

func ErrorHandler(c fiber.Ctx, err error) error {
	// Check if the error is an AppError
	if IsAppError(err) {
		e := err.(*AppError)
		if err := c.Status(e.Code).JSON(e.toResponse()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "A more critical error occurred.",
					"status":  StatusInternalServerError,
				},
			})
		}
		return nil
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		if err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"message": fiberErr.Error(),
				"status":  StatusFiberError,
			},
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "A more critical error occurred.",
					"status":  StatusInternalServerError,
				},
			})
		}
		return nil
	}

	// For all other errors, return a generic internal server error response
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{
			"message": "Unhandled error.",
			"status":  StatusInternalServerError,
		},
	})
}

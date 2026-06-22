package apperror

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"runtime"

	"github.com/alkuinvito/ai-assistant/pkg/logger"
	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/sirupsen/logrus"
)

type AppError struct {
	context    context.Context
	err        error
	data       any
	message    string
	statusCode int
	traceId    string
}

type ErrorField struct {
	Key   string
	Value any
}

func New(ctx context.Context, statusCode int, message string, err error, errorFields ...*ErrorField) *AppError {
	pc, file, line, _ := runtime.Caller(2)
	trace := fmt.Sprintf("%s:%d", path.Base(file), line)
	caller := path.Base(runtime.FuncForPC(pc).Name())

	errorData := make(map[string]any)
	for _, field := range errorFields {
		errorData[field.Key] = field.Value
	}

	traceId := utils.GetTraceID(ctx)

	logger.Logger().WithError(err).WithFields(logrus.Fields{
		"trace_id": traceId,
		"trace":    trace,
		"caller":   caller,
		"data":     errorData,
	}).Error(message)

	return &AppError{
		traceId:    traceId,
		statusCode: statusCode,
		message:    message,
		err:        errors.New(message),
	}
}

func NewBadRequest(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusBadRequest, message, err, fields...)
}

func NewConflict(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusConflict, message, err, fields...)
}

func NewForbidden(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusForbidden, message, err, fields...)
}

func NewInternalServerError(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusInternalServerError, message, err, fields...)
}

func NewNotFound(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusNotFound, message, err, fields...)
}

func NewUnauthorized(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusUnauthorized, message, err, fields...)
}

func NewUnprocessableEntity(ctx context.Context, message string, err error, fields ...*ErrorField) *AppError {
	return New(ctx, http.StatusUnprocessableEntity, message, err, fields...)
}

func (ae *AppError) Error() string {
	return ae.err.Error()
}

func (ae *AppError) Message() string {
	return ae.message
}

func (ae *AppError) StatusCode() int {
	return ae.statusCode
}

func (ae *AppError) Unwrap() error {
	return ae.err
}

func Field(key string, value any) *ErrorField {
	value = utils.Stringify(value)
	return &ErrorField{Key: key, Value: value}
}

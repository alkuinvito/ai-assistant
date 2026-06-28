package service_error

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

type ServiceError struct {
	err        error
	message    string
	statusCode int
	requestId  string
}

type ErrorField struct {
	Key   string
	Value any
}

func New(ctx context.Context, statusCode int, message string, err error, errorFields ...*ErrorField) *ServiceError {
	pc, file, line, _ := runtime.Caller(2)
	trace := fmt.Sprintf("%s:%d", path.Base(file), line)
	caller := path.Base(runtime.FuncForPC(pc).Name())

	errorData := make(map[string]any, len(errorFields))
	for _, field := range errorFields {
		errorData[field.Key] = field.Value
	}

	requestId := utils.GetRequestID(ctx)
	logger.Logger().WithError(err).WithFields(logrus.Fields{
		"req_id": requestId,
		"trace":  trace,
		"caller": caller,
		"data":   errorData,
	}).Error(message)

	return &ServiceError{
		err:        errors.New(message),
		message:    message,
		requestId:  requestId,
		statusCode: statusCode,
	}
}

func NewBadRequest(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusBadRequest, message, err, fields...)
}

func NewConflict(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusConflict, message, err, fields...)
}

func NewForbidden(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusForbidden, message, err, fields...)
}

func NewInternalServerError(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusInternalServerError, message, err, fields...)
}

func NewNotFound(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusNotFound, message, err, fields...)
}

func NewTooManyRequests(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusTooManyRequests, message, err, fields...)
}

func NewUnauthorized(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusUnauthorized, message, err, fields...)
}

func NewUnprocessableEntity(ctx context.Context, message string, err error, fields ...*ErrorField) *ServiceError {
	return New(ctx, http.StatusUnprocessableEntity, message, err, fields...)
}

func (ae *ServiceError) Error() string {
	return ae.err.Error()
}

func (ae *ServiceError) Message() string {
	return ae.message
}

func (ae *ServiceError) StatusCode() int {
	return ae.statusCode
}

func (ae *ServiceError) Unwrap() error {
	return ae.err
}

func Field(key string, value any) *ErrorField {
	value = utils.Stringify(value)
	return &ErrorField{Key: key, Value: value}
}

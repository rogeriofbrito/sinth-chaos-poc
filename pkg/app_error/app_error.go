package app_error

import "errors"

type AppError string

const (
	FAULT_ENGINE AppError = "missing value"
	BINARY       AppError = "missing value"
)

func New(err AppError) error {
	return errors.New(string(err))
}

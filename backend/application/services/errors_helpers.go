package services

import (
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
)

func validation(field string, value interface{}, message string) error {
	return domainerrors.NewValidationError(field, value, message)
}

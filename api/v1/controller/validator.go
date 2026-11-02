package controller

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validateStruct(input interface{}) error {
	if err := validate.Struct(input); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("data tidak valid")
		}

		messages := make([]string, 0, len(validationErrors))
		for _, fe := range validationErrors {
			messages = append(messages, formatFieldError(fe))
		}
		return fmt.Errorf(strings.Join(messages, "; "))
	}
	return nil
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return fmt.Sprintf("%s harus berupa format email yang valid", field)
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s tidak valid (%s)", field, fe.Tag())
	}
}
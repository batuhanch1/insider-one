package notification_management_api

import (
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func gtNow(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(time.Time)

	if field.IsZero() {
		return true
	}

	return field.After(time.Now())
}

func withinOneYear(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(time.Time)

	if field.IsZero() {
		return true
	}

	return field.Before(time.Now().AddDate(1, 0, 0))
}

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("gt_now", gtNow)
		v.RegisterValidation("within_one_year", withinOneYear)
	}
}

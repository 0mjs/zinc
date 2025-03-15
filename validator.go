package zinc

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var errors []string
	for _, err := range ve {
		errors = append(errors, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(errors, "; ")
}

type Validator struct {
	tagName string
}

func NewValidator() *Validator {
	return &Validator{
		tagName: "validate",
	}
}

func (v *Validator) Validate(s interface{}) ValidationErrors {
	var errors ValidationErrors
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		tag := fieldType.Tag.Get(v.tagName)
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if err := v.validateField(field, fieldType.Name, rule); err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldType.Name,
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

func (v *Validator) validateField(field reflect.Value, fieldName, rule string) error {
	parts := strings.Split(rule, "=")
	ruleName := parts[0]
	ruleValue := ""
	if len(parts) > 1 {
		ruleValue = parts[1]
	}

	switch ruleName {
	case "required":
		if isZero(field) {
			return fmt.Errorf("field is required")
		}
	case "email":
		if !isValidEmail(field.String()) {
			return fmt.Errorf("invalid email format")
		}
	case "min":
		min, _ := strconv.Atoi(ruleValue)
		if !validateMin(field, min) {
			return fmt.Errorf("value must be at least %d", min)
		}
	case "max":
		max, _ := strconv.Atoi(ruleValue)
		if !validateMax(field, max) {
			return fmt.Errorf("value must be at most %d", max)
		}
	case "len":
		length, _ := strconv.Atoi(ruleValue)
		if !validateLength(field, length) {
			return fmt.Errorf("length must be %d", length)
		}
	}

	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func validateMin(field reflect.Value, min int) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(min)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(min)
	case reflect.String:
		return len(field.String()) >= min
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() >= min
	default:
		return true
	}
}

func validateMax(field reflect.Value, max int) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(max)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(max)
	case reflect.String:
		return len(field.String()) <= max
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() <= max
	default:
		return true
	}
}

func validateLength(field reflect.Value, length int) bool {
	switch field.Kind() {
	case reflect.String:
		return len(field.String()) == length
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == length
	default:
		return true
	}
}

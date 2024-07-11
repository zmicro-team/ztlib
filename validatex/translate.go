package validatex

import (
	"errors"
	"strings"

	zhLang "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
)

var defaultZhTrans = func() ut.Translator {
	z := zhLang.New()
	uni := ut.New(z, z)
	trans, b := uni.GetTranslator("zh")
	if !b {
		panic("not zh")
	}
	return trans
}()

var DefaultZhTrans = defaultZhTrans

func RegisterDefaultTranslations(valid *validator.Validate) error {
	err := zh.RegisterDefaultTranslations(valid, defaultZhTrans)
	return err
}

type TranslateError struct{}

// Translate translates validation errors to a localized error message.
// It takes an error as input and returns a translated error.
// If the input error is not of type validator.ValidationErrors, it returns the error as is.
// Otherwise, it translates each validation error using the defaultZhTrans translator and
// appends the translated error messages to a slice. Finally, it joins the error messages
// with newline characters and returns a new error containing the joined messages.
func (t TranslateError) Translate(err error) error {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	detail := []string{}
	for k, _ := range errs {
		item := errs[k]
		detail = append(detail, item.Translate(defaultZhTrans))
	}
	return errors.New(strings.Join(detail, "\n"))
}

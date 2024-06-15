package validatex

import (
	"reflect"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

func RegisterDefaultValidators(v *validator.Validate, trans ut.Translator) error {
	return Register(defaultFieldValidators, v, trans)
}

// Register 函数用于注册字段验证器和翻译器到 validator.Validate 实例中。
// 参数：
//   - validators: 字段验证器的列表，实现了 IFieldValidator 接口。
//   - v: validator.Validate 实例，用于注册验证器。
//   - trans: ut.Translator 实例，用于注册翻译器。
// 返回值：
//   - err: 错误信息，如果注册过程中发生错误，则返回相应的错误。
func Register(validators []IFieldValidator, v *validator.Validate, trans ut.Translator) (err error) {
	v.RegisterCustomTypeFunc(
		func(field reflect.Value) any {
			if valuer, ok := field.Interface().(decimal.Decimal); ok {
				return valuer.String()
			}
			return nil
		},
		decimal.Decimal{},
	)

	if len(validators) == 0 {
		return
	}

	// * register validation
	for _, t := range validators {
		err = v.RegisterValidation(t.Tag(), t.RegisterFun())
		if err != nil {
			return err
		}
	}

	// * register translation
	if trans != nil {

		for _, t := range validators {

			if t.CustomTransFunc() != nil && t.CustomRegisFunc() != nil {

				err = v.RegisterTranslation(t.Tag(), trans, t.CustomRegisFunc(), t.CustomTransFunc())

			} else if t.CustomTransFunc() != nil && t.CustomRegisFunc() == nil {

				err = v.RegisterTranslation(t.Tag(), trans, registrationFunc(t.Tag(), t.Translation(), t.Override()), t.CustomTransFunc())

			} else if t.CustomTransFunc() == nil && t.CustomRegisFunc() != nil {

				err = v.RegisterTranslation(t.Tag(), trans, t.CustomRegisFunc(), translateFunc)

			} else {
				err = v.RegisterTranslation(t.Tag(), trans, registrationFunc(t.Tag(), t.Translation(), t.Override()), translateFunc)
			}

			if err != nil {
				return
			}
		}

	}

	return
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, override); err != nil {
			return
		}
		return
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field())
	if err != nil {
		return fe.(error).Error()
	}
	return t
}

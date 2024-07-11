package validatex

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var defaultFieldValidators = []IFieldValidator{
	NewFieldValidator(
		SetTag("mobile"),
		SetTranslation("{0}格式不正确,请输入正确的手机格式"),
		SetRegisterFun(ValidateIsMobile),
	),
	NewFieldValidator(
		SetTag("uniform_code"),
		SetTranslation("{0}格式不正确,请输入正确的统一社会信用代码"),
		SetRegisterFun(ValidateIsUniformCode),
	),
	NewFieldValidator(
		SetTag("id_card"),
		SetTranslation("{0}格式不正确,请输入正确的身份证号码"),
		SetRegisterFun(ValidateIsIDCard),
	),
	NewFieldValidator(
		SetTag("iso8601"),
		SetTranslation("{0}格式不正确,请输入正确的ISO8601时间格式"),
		SetRegisterFun(ValidateIsISO8601),
	),
	NewFieldValidator(
		SetTag("ngte"),
		// SetTranslation("字符串{0}必须大于等于"),
		SetRegisterFun(ValidateNumericGte),
		SetCustomRegisFunc(func(ut ut.Translator) (err error) {
			if err = ut.Add("ngte-string", "数字字符串{0}必须大于等于{1}", false); err != nil {
				return
			}
			return
		}),
		SetCustomTransFunc(func(ut ut.Translator, fe validator.FieldError) string {
			var err error
			var t string

			var digits uint64
			var kind reflect.Kind

			if idx := strings.Index(fe.Param(), "."); idx != -1 {
				digits = uint64(len(fe.Param()[idx+1:]))
			}

			f64, err := strconv.ParseFloat(fe.Param(), 64)
			if err != nil {
				goto END
			}

			kind = fe.Kind()
			if kind == reflect.Ptr {
				kind = fe.Type().Elem().Kind()
			}

			switch kind {
			case reflect.String:
				t, err = ut.T("ngte-string", fe.Field(), ut.FmtNumber(f64, digits))
			default:
				return fe.(error).Error()
			}

		END:
			if err != nil {
				fmt.Printf("警告: 翻译字段错误: %s\n", err)
				return fe.(error).Error()
			}

			return t
		}),
	),
	NewFieldValidator(
		SetTag("nlte"),
		// SetTranslation(),
		SetRegisterFun(ValidateNumericLte),
		SetCustomRegisFunc(func(ut ut.Translator) (err error) {
			if err = ut.Add("nlte-string", "数字字符串{0}必须小于等于{1}", false); err != nil {
				return
			}
			return
		}),
		SetCustomTransFunc(func(ut ut.Translator, fe validator.FieldError) string {
			var err error
			var t string

			var digits uint64
			var kind reflect.Kind

			if idx := strings.Index(fe.Param(), "."); idx != -1 {
				digits = uint64(len(fe.Param()[idx+1:]))
			}

			f64, err := strconv.ParseFloat(fe.Param(), 64)
			if err != nil {
				goto END
			}

			kind = fe.Kind()
			if kind == reflect.Ptr {
				kind = fe.Type().Elem().Kind()
			}

			switch kind {
			case reflect.String:
				t, err = ut.T("nlte-string", fe.Field(), ut.FmtNumber(f64, digits))
			default:
				return fe.(error).Error()
			}

		END:
			if err != nil {
				fmt.Printf("警告: 翻译字段错误: %s\n", err)
				return fe.(error).Error()
			}

			return t
		}),
	),
}

package validatex

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/zmicro-team/ztlib/validatex/idvalidator"
	"github.com/zmicro-team/ztlib/validatex/uniform"
)

// 是否为手机
var rxMobile = regexp.MustCompile(`^1[3456789]\d{9}$`)

// 是否为ISO8601时间格式
var iso8601 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?([+-]\d{2}:?\d{2}|Z)$`)

// \b[A-Z]\d{6}[\(][(A-Z0-9][\)]
// OR
// \b[A-Z][A-Z]\d{6}\(\d\)
// 香港身份证
var rxHkIdCard = regexp.MustCompile(`\b[A-Z]\d{6}[\(][(A-Z0-9][\)]|\b[A-Z][A-Z]\d{6}\(\d\)`)

func ValidateIsMobile(fl validator.FieldLevel) bool {
	return rxMobile.MatchString(fl.Field().String())
}

// * 统一信用代码
func ValidateIsUniformCode(fl validator.FieldLevel) bool {
	return uniform.CalibrationUniform321002015(fl.Field().String())
}

// * 身份证号
func ValidateIsIDCard(fl validator.FieldLevel) bool {
	if idvalidator.IsValidCitizenNo(fl.Field().String()) || rxHkIdCard.MatchString(fl.Field().String()) {
		return true
	}
	return false
}

// 是否为ISO8601时间格式
func ValidateIsISO8601(fl validator.FieldLevel) bool {
	return iso8601.MatchString(fl.Field().String())
}

// fl >= where
// * 字符串数字大于等于
func ValidateNumericGte(fl validator.FieldLevel) bool {
	n, err := decimal.NewFromString(fl.Field().String())
	if err != nil {
		return false
	}
	p, err := decimal.NewFromString(strings.TrimSpace(fl.Param()))
	if err != nil {
		return false
	}
	return n.GreaterThanOrEqual(p)
}

// * 字符串数字小于等于
func ValidateNumericLte(fl validator.FieldLevel) bool {
	n, err := decimal.NewFromString(fl.Field().String())
	if err != nil {
		return false
	}
	p, err := decimal.NewFromString(strings.TrimSpace(fl.Param()))
	if err != nil {
		return false
	}
	return n.LessThanOrEqual(p)
}

// * 香港身份证
func ValidateIsHKIDCard(fl validator.FieldLevel) bool {
	return rxHkIdCard.MatchString(fl.Field().String())
}

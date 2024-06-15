package validatex

import "github.com/go-playground/validator/v10"

type IFieldValidator interface {
	Tag() string
	Translation() string
	Override() bool
	CustomRegisFunc() validator.RegisterTranslationsFunc
	CustomTransFunc() validator.TranslationFunc
	RegisterFun() func(validator.FieldLevel) bool
}

type FieldValidator struct {
	tag             string
	translation     string
	override        bool
	customRegisFunc validator.RegisterTranslationsFunc
	customTransFunc validator.TranslationFunc
	registerFun     func(validator.FieldLevel) bool
}

type SetFieldValidatorOption func(field *FieldValidator)

func SetTag(tag string) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.tag = tag
	}
}

func SetTranslation(translation string) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.translation = translation
	}
}

func SetOverride(override bool) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.override = override
	}
}

func SetCustomRegisFunc(customRegisFunc validator.RegisterTranslationsFunc) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.customRegisFunc = customRegisFunc
	}
}

func SetCustomTransFunc(customTransFunc validator.TranslationFunc) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.customTransFunc = customTransFunc
	}
}

func SetRegisterFun(registerFun func(validator.FieldLevel) bool) SetFieldValidatorOption {
	return func(field *FieldValidator) {
		field.registerFun = registerFun
	}
}

func NewFieldValidator(options ...SetFieldValidatorOption) IFieldValidator {
	field := &FieldValidator{
		override:        false,
		customRegisFunc: nil,
		customTransFunc: nil,
	}
	for _, option := range options {
		option(field)
	}
	if field.registerFun == nil {
		panic("registerFun is nil")
	}
	if field.tag == "" {
		panic("tag is empty")
	}
	if field.translation == "" {
		panic("translation is empty")
	}
	return field
}

func (field FieldValidator) Tag() string {
	return field.tag
}

func (field FieldValidator) Translation() string {
	return field.translation
}

func (field FieldValidator) Override() bool {
	return field.override
}

func (field FieldValidator) CustomRegisFunc() validator.RegisterTranslationsFunc {
	return field.customRegisFunc
}

func (field FieldValidator) CustomTransFunc() validator.TranslationFunc {
	return field.customTransFunc
}

func (field FieldValidator) RegisterFun() func(validator.FieldLevel) bool {
	return field.registerFun
}

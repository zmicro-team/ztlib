package validatex

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
		SetTranslation("字符串{0}必须大于等于{1}"),
		SetRegisterFun(ValidateNumericGte),
	),
	NewFieldValidator(
		SetTag("nlte"),
		SetTranslation("字符串{0}必须小于等于{1}"),
		SetRegisterFun(ValidateNumericLte),
	),
	
}

package idvalidator

import (
	"errors"
	"strconv"
	"time"
)

// 定义全局变量
var (
	// 身份证前17位的权重系数
	weight = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

	// 校验码对应值
	validValue = [11]byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	// 省份代码映射表
	validProvince = map[string]string{
		"11": "北京市",
		"12": "天津市",
		"13": "河北省",
		"14": "山西省",
		"15": "内蒙古自治区",
		"21": "辽宁省",
		"22": "吉林省",
		"23": "黑龙江省",
		"31": "上海市",
		"32": "江苏省",
		"33": "浙江省",
		"34": "安徽省",
		"35": "福建省",
		"36": "江西省",
		"37": "山东省",
		"41": "河南省",
		"42": "湖北省",
		"43": "湖南省",
		"44": "广东省",
		"45": "广西壮族自治区",
		"46": "海南省",
		"50": "重庆市",
		"51": "四川省",
		"52": "贵州省",
		"53": "云南省",
		"54": "西藏自治区",
		"61": "陕西省",
		"62": "甘肃省",
		"63": "青海省",
		"64": "宁夏回族自治区",
		"65": "新疆维吾尔自治区",
		"71": "台湾省",
		"81": "香港特别行政区",
		"91": "澳门特别行政区",
	}
)

// IsValid 验证身份证号码是否有效(校验码检查)
func IsValid(citizenNo string) bool {
	// 检查长度是否为18位
	if len(citizenNo) != 18 {
		return false
	}

	sum := 0
	// 计算前17位的加权和
	for i := 0; i < 17; i++ {
		n, err := strconv.Atoi(string(citizenNo[i]))
		if err != nil {
			return false
		}
		sum += n * weight[i]
	}

	// 计算校验码
	mod := sum % 11
	return validValue[mod] == citizenNo[17]
}

// isLeapYear 判断是否为闰年
func isLeapYear(year int) bool {
	if year <= 0 {
		return false
	}
	// 闰年规则：能被4整除但不能被100整除，或者能被400整除
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// CheckBirthdayValid 验证身份证中的出生日期是否有效
func CheckBirthdayValid(year, month, day int) bool {
	// 基本日期范围检查
	if year < 1900 || month <= 0 || month > 12 || day <= 0 || day > 31 {
		return false
	}

	// 检查是否未来日期
	now := time.Now()
	if year > now.Year() ||
		(year == now.Year() && month > int(now.Month())) ||
		(year == now.Year() && month == int(now.Month()) && day > now.Day()) {
		return false
	}

	// 检查各月份的天数
	switch month {
	case 2: // 二月特殊处理
		if isLeapYear(year) {
			if day > 29 {
				return false
			}
		} else if day > 28 {
			return false
		}
	case 4, 6, 9, 11: // 30天的月份
		if day > 30 {
			return false
		}
	}

	return true
}

// CheckProvinceValid 检查省份代码是否有效
func CheckProvinceValid(citizenNo string) bool {
	if len(citizenNo) < 2 {
		return false
	}
	// 直接从省份代码映射表中查找
	_, exists := validProvince[citizenNo[:2]]
	return exists
}

// IsValidCitizenNo 综合验证身份证号码的有效性
func IsValidCitizenNo(citizenNo string) bool {
	// 1. 校验码检查
	if !IsValid(citizenNo) {
		return false
	}

	// 2. 检查所有字符是否合法
	for i := 0; i < 17; i++ {
		if citizenNo[i] < '0' || citizenNo[i] > '9' {
			return false
		}
	}
	// 第18位可以是X或数字
	if !(citizenNo[17] == 'X' || (citizenNo[17] >= '0' && citizenNo[17] <= '9')) {
		return false
	}

	// 3. 检查省份代码
	if !CheckProvinceValid(citizenNo) {
		return false
	}

	// 4. 检查出生日期
	year, err := strconv.Atoi(citizenNo[6:10])
	if err != nil {
		return false
	}
	month, err := strconv.Atoi(citizenNo[10:12])
	if err != nil {
		return false
	}
	day, err := strconv.Atoi(citizenNo[12:14])
	if err != nil {
		return false
	}

	return CheckBirthdayValid(year, month, day)
}

// GetCitizenNoInfo 从有效身份证号码中提取信息
func GetCitizenNoInfo(citizenNo string) (birthday time.Time, sex string, address string, err error) {
	// 先验证身份证号码是否有效
	if !IsValidCitizenNo(citizenNo) {
		err = errors.New("无效的身份证号码")
		return
	}

	// 解析出生日期(格式:2006-01-02)
	birthday, err = time.Parse("2006-01-02",
		citizenNo[6:10]+"-"+citizenNo[10:12]+"-"+citizenNo[12:14])
	if err != nil {
		return
	}

	// 根据第17位判断性别(奇数为男，偶数为女)
	genderDigit, _ := strconv.Atoi(string(citizenNo[16]))
	if genderDigit%2 == 0 {
		sex = "女"
	} else {
		sex = "男"
	}

	// 根据前两位获取省份信息
	address = validProvince[citizenNo[:2]]

	return
}

func GetProvince(code string) string {
	return validProvince[code[:2]]
}

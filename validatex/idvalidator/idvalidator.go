package idvalidator

import (
	"bytes"
	"errors"
	"strconv"
	"time"
)

var weight = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var validValue = [11]byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
var validProvince = map[string]string{
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
	"36": "山西省",
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

func IsValid(citizenNo string) bool {
	nLen := len(citizenNo)
	if nLen != 18 {
		return false
	}

	nSum := 0
	for i := 0; i < nLen-1; i++ {
		n, _ := strconv.Atoi(string(citizenNo[i]))
		nSum += n * weight[i]
	}
	mod := nSum % 11
	if validValue[mod] == citizenNo[17] {
		return true
	}

	return false
}

func isLeapYear(nYear int) bool {
	if nYear <= 0 {
		return false
	}

	if (nYear%4 == 0 && nYear%100 != 0) || nYear%400 == 0 {
		return true
	}

	return false
}

func CheckBirthdayValid(nYear, nMonth, nDay int) bool {
	if nYear < 1900 || nMonth <= 0 || nMonth > 12 || nDay <= 0 || nDay > 31 {
		return false
	}

	curYear, curMonth, curDay := time.Now().Date()
	if nYear == curYear {
		if nMonth > int(curMonth) {
			return false
		} else if nMonth == int(curMonth) && nDay > curDay {
			return false
		}
	}

	if 2 == nMonth {
		if isLeapYear(nYear) {
			if nDay > 29 {
				return false
			}
		} else {
			if nDay > 28 {
				return false
			}
		}
	} else if 4 == nMonth || 6 == nMonth || 9 == nMonth || 11 == nMonth {
		if nDay > 30 {
			return false
		}
	}

	return true
}

func CheckProvinceValid(citizenNo string) bool {
	provinceCode := make([]byte, 0)
	provinceCode = append(provinceCode, citizenNo[:2]...)
	provinceStr := string(provinceCode)

	for i := range validProvince {
		if provinceStr == i {
			return true
		}
	}

	return false
}

func IsValidCitizenNo(citizenNo string) bool {

	if !IsValid(citizenNo) {
		return false
	}

	for i, v := range citizenNo {
		n, _ := strconv.Atoi(string(v))
		if n >= 0 && n <= 9 {
			continue
		}

		if v == 'X' && i == 16 {
			continue
		}

		return false
	}

	if !CheckProvinceValid(citizenNo) {
		return false
	}

	nYear, _ := strconv.Atoi(citizenNo[6:10])
	nMonth, _ := strconv.Atoi(citizenNo[10:12])
	nDay, _ := strconv.Atoi(citizenNo[12:14])
	if !CheckBirthdayValid(nYear, nMonth, nDay) {
		return false
	}

	return true
}

func GetCitizenNoInfo(citizenNo string) (birthday time.Time, sex string, address string, err error) {
	err = nil
	if !IsValidCitizenNo(citizenNo) {
		err = errors.New("不合法的身份证号码")
		return
	}

	buf := bytes.Buffer{}
	buf.WriteString(citizenNo[6:10])
	buf.WriteString("-")
	buf.WriteString(citizenNo[10:12])
	buf.WriteString("-")
	buf.WriteString(citizenNo[12:14])

	birthday, _ = time.Parse("2006-01-02", buf.String())
	sex = getSex(citizenNo[16])
	address = validProvince[citizenNo[:2]]

	return
}

func getSex(s byte) string {
	var sex string
	genderMask, _ := strconv.Atoi(string(s))
	if genderMask%2 == 0 {
		sex = "女"
	} else {
		sex = "男"
	}
	return sex
}

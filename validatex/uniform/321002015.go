package uniform

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	/*
		1、统一社会信用代码由18位数字或大写英文字母（不使用I、O、Z、S、V）组成，不含有汉字。
		2、第1位：登记管理部门代码（共一位字符）。
		分为 1机构编制；2外交；3司法行政；4文化；5民政；6旅游；7宗教；8工会；9工商；A中央军委改革和编制办公室；N农业；Y其他。
	*/

	registrationAuthorityCode = []string{
		"1", // 机构编制
		"2", // 外交
		"3", // 司法行政
		"4", // 文化
		"5", // 民政
		"6", // 旅游
		"7", // 宗教
		"8", // 工会
		"9", // 工商
		"A", // 中央军委改革和编制办公室
		"N", // 农业
		"Y", // 其他
	}

	/*
		 # 第2位：机构类别代码（共一位字符）。
		　 第2位：机构类别代码，使用阿拉伯数字表示。分为：
		　　1机构编制：1机关，2事业单位，3中央编办直接管理机构编制的群众团体，9其他；
		　　2外交：1外国常住新闻机构，9其他；
		　　3司法行政：1律师执业机构，2公证处，3基层法律服务所，4司法鉴定机构，5仲裁委员会，9其他；
		　　4文化：1外国在华文化中心，9其他；
		　　5民政：1社会团体，2民办非企业单位，3基金会，9其他；
		　　6旅游：1外国旅游部门常驻代表机构，2港澳台地区旅游部门常驻内地（大陆）代表机构，9其他；7宗教：1宗教活动场所，2宗教院校，9其他；
		　　8工会：1基层工会，9其他；
		　　9工商：1企业，2个体工商户，3农民专业合作社；
		　　A中央军委改革和编制办公室：1军队事业单位，9其他；
		　　N农业：1组级集体经济组织，2村级集体经济组织，3乡镇级集体经济组织，9其他；
		　　Y其他：不再具体划分机构类别，统一用1表示。
	*/
	organizationCategoryCode = map[string][]string{
		"1": {"1", "2", "3", "9"},
		"2": {"1", "9"},
		"3": {"1", "2", "3", "4", "5", "9"},
		"4": {"1", "9"},
		"5": {"1", "2", "3", "9"},
		"6": {"1", "2", "9"},
		"7": {"1", "2", "9"},
		"8": {"1", "9"},
		"9": {"1", "2", "3"},
		"A": {"1", "9"},
		"N": {"1", "2", "3", "9"},
		"Y": {"1"},
	}
	// 4、第3位~第8位：登记管理机关行政区划码（共6位阿拉伯数字）。
	// 5、第9位~第17位：主体标识码（组织机构代码）（共9位字符）。
	// 6、第18位：校验码（共一位字符）。
	// 7、统一社会信用代码采用ISO 7064:2003，MOD 31-3的校验码系统。

	// 统一社会信用代码可用字符 不含I、O、S、V、Z
	codeOrigin = "0123456789ABCDEFGHJKLMNPQRTUWXY"

	// 统一社会信用代码相对应顺序的加权因子
	weightedfactors = []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}

	// 正则
	uniform321002015Regex = regexp.MustCompile(`^[1-9A-HJ-NP-RT-Y]\d{1}[0-9A-HJ-NP-RT-Y]{6}[0-9A-HJ-NP-RT-Y]{9}[0-9A-HJ-NP-RT-Y]$`)
)

type Uniform321002015 struct {
	Rd  string // * 登记部门[0]
	Rc  string // * 机构类别(1-9,A,N,Y)[1]
	Ad  string // * 行政区划[2-7]
	Oc  string // * 主体标识码(组织机构代码)[8-16]
	Cc  string // * 校验码[17]
	Org string // * 完整代码
}

/*
  创建统一社会信用代码
*/
func NewUniform321002015(code string) (u *Uniform321002015, err error) {
	if len(code) != 18 {
		return nil, fmt.Errorf("invalid lenght: %d", len(code))
	}
	checkCode, err := calculateCheckCode(code)
	if err != nil {
		return nil, err
	}
	if checkCode != string(code[17]) {
		return nil, fmt.Errorf("invalid calibration code: %s", code)
	}
	return &Uniform321002015{
		Rd:  code[0:1],
		Rc:  code[1:2],
		Ad:  code[2:8],
		Oc:  code[8:17],
		Cc:  code[17:18],
		Org: code,
	}, nil
}

/*
	  校验统一社会信用代码(完整校验)
*/
func CalibrationUniform321002015(code string) bool {
	if len(code) != 18 {
		return false
	}
	checkCode, err := calculateCheckCode(code)
	if err != nil {
		return false
	}
	if checkCode != string(code[17]) {
		return false
	}
	return true
}

func Uniform321002015Regex(code string) bool {
	return uniform321002015Regex.MatchString(code)
}

/*
	计算校验码
	MOD 31
*/
func calculateCheckCode(code string) (string, error) {
	sum := 0
	for i, char := range code[:17] {
		index := strings.Index(codeOrigin, string(char))
		if index == -1 {
			return "", fmt.Errorf("char %s is not in codeOrigin: %s", string(char), code)
		}
		weight := weightedfactors[i]
		sum += index * weight
	}
	remainder := sum % 31
	checkCode := (31 - remainder) % 31
	if checkCode > len(codeOrigin)-1 {
		return "", fmt.Errorf("length of codeOrigin is not enough: %s", code)
	}
	return string(codeOrigin[checkCode]), nil
}

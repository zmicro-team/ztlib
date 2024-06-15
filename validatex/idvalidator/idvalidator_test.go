package idvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCitizenNoInfo(t *testing.T) {
	testdata := []struct {
		Sex     string
		Address string
		IdNo    string
		Birth   string
	}{
		{
			Sex:     "男",
			Address: "北京市",
			IdNo:    "110101198001010010",
			Birth:   "1980-01-01",
		},
		{
			Sex:     "女",
			Address: "江苏省",
			IdNo:    "320102198001010024",
			Birth:   "1980-01-01",
		},
		{
			Sex:     "女",
			Address: "湖北省",
			IdNo:    "420112200507090048",
			Birth:   "2005-07-09",
		},
	}

	for _, data := range testdata {
		birth, sex, address, err := GetCitizenNoInfo(data.IdNo)
		assert.NoError(t, err)
		assert.Equal(t, data.Sex, sex)
		assert.Equal(t, data.Address, address)
		assert.Equal(t, data.Birth, birth.Format("2006-01-02"))
	}
}

func TestIsValid(t *testing.T) {
	vaildData := []string{
		"420112200507090208",
		"420112200507090224",
		"420112200507090240",
		"420112200507090267",
		"420112200507090283",
		"420112200507090304",
		"340101200907090224",
		"340101200907090240",
		"340101200907090267",
		"340101200907090283",
	}

	invaildData := []string{
		"340101200907091420",
		"340101200907091440",
		"340101200907091460",
		"340101200907091660",
		"340101200907091680",
		"340101200907091703",
		"340101200907091723",
		"340101200907091743",
		"340101200907091763",
		"340101200907091963",
	}

	for i := 0; i < 10; i++ {
		assert.True(t, IsValid(vaildData[i]))
		assert.False(t, IsValid(invaildData[i]))
		assert.True(t, IsValidCitizenNo(vaildData[i]))
		assert.False(t, IsValidCitizenNo(invaildData[i]))
	}

}

func BenchmarkIsValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsValid("340101200907090283")
	}
}

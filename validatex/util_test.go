package validatex

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/zmicro-team/ztlib/validatex/idvalidator"
	"github.com/zmicro-team/ztlib/validatex/uniform"
)

func TestUncode(t *testing.T) {
	assert.True(t, uniform.CalibrationUniform321002015("911201143858628820"))
}

func TestIdCard(t *testing.T) {
	assert.True(t, idvalidator.IsValidCitizenNo("421321200502045772"))
}

func TestValidateIsHKIDCard(t *testing.T) {
	assert.True(t, rxHkIdCard.MatchString("C668668(1)"))
	assert.True(t, rxHkIdCard.MatchString("R458631(9)"))
}

func TestValidateIsMobile(t *testing.T) {
	assert.True(t, rxMobile.MatchString("15892101012"))
}

func TestValidateIsISO8601(t *testing.T) {
	assert.True(t, iso8601.MatchString("2020-01-01T00:00:00Z"))
}

func TestRegisterDefaultValidators(t *testing.T) {
	valid := validator.New()
	err := RegisterDefaultValidators(valid, DefaultZhTrans)
	assert.Nil(t, err)
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"mobile", "15892101012", true},
		{"uniform_code", "91320507MA21XXFU2A", true},
		{"id_card", "110101198001010010", true},
		{"iso8601", "2020-01-01T00:00:00Z", true},
		{"ngte=10", "10", true},
		{"required,numeric,ngte=0", "1", true},
		{"nlte=10", "10", true},
		{"nlte=110", "9", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := valid.Var(tt.value, tt.name)
			if err != nil {
				t.Log(err.Error())
			}
			assert.Nil(t, err)
		})
	}
}

func TestRegisterT(t *testing.T) {
	valid := validator.New()
	valid.SetTagName("binding")
	RegisterDefaultTranslations(valid)
	err := RegisterDefaultValidators(valid, DefaultZhTrans)
	assert.Nil(t, err)
	type RegisterBillSupportWithdrawRequest struct {
		// @gotags: binding:"required,numeric,ngte=1" form:"amt" comment:"金额"
		Amt string `protobuf:"bytes,2,opt,name=amt,proto3" json:"amt,omitempty" binding:"required,numeric,ngte=1" form:"amt" comment:"金额"`
		// @gotags: binding:"required,numeric,ngte=-1" form:"fee" comment:"手续费"
		Fee string `protobuf:"bytes,3,opt,name=fee,proto3" json:"fee,omitempty" binding:"required,numeric,ngte=0" form:"fee" comment:"手续费"`
	}
	req := &RegisterBillSupportWithdrawRequest{
		Amt: "-1",
		Fee: "-1",
	}
	err = valid.Struct(req)
	ts := TranslateError{}
	err = ts.Translate(err)
	assert.NotEmpty(t, err)
	if err != nil {
		t.Log(err.Error())
	}
}

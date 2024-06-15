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
		{"ngte=9", "10", true},
		{"nlte=10", "10", true},
		{"nlte=110", "9", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := valid.Var(tt.value, tt.name)
			assert.Nil(t, err)
		})
	}
}

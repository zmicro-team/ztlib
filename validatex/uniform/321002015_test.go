package uniform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCheckCode(t *testing.T) {
	tests := []struct {
		Code      string
		CheckCode string
		Ok        bool
	}{
		{"91350100M000100Y43", "3", true},
		{"91350203MA31F1331W", "W", true},
		{"91350213MA33WMAT03", "3", true},
		{"12350200426607600N", "N", true},
		{"91320507MA21XXFU2A", "A", true},
		{"91320118MA21R1P51X", "X", true},
		{"92321283MA21R1P43M", "M", true},
		{"91350822MA34D63G3C", "C", true},
		{"91350000070893203F", "F", true},
		{"91320507MA21XXFU2A", "8", false},
	}

	for i := 0; i < len(tests); i++ {
		resulst, err := calculateCheckCode(tests[i].Code)
		if tests[i].Ok {
			assert.Equal(t, err, nil)
			assert.Equal(t, tests[i].CheckCode, resulst)
		} else {
			assert.NotEqual(t, tests[i].CheckCode, resulst)
		}
	}
}

func TestUniform321002015Regex(t *testing.T) {
	tests := []struct {
		Code      string
		CheckCode string
		Ok        bool
	}{
		{"91350100M000100Y43", "3", true},
		{"91350203MA31F1331W", "W", true},
		{"91350213MA33WMAT03", "3", true},
		{"12350200426607600N", "N", true},
		{"91320507MA21XXFU2A", "A", true},
		{"91320118MA21R1P51X", "X", true},
		{"92321283MA21R1P43M", "M", true},
		// * 只匹配格式, 不校验校验码
		{"91320507MA21XXFU2A", "8", true},
	}

	for i := 0; i < len(tests); i++ {
		assert.Equal(t, Uniform321002015Regex(tests[i].Code), tests[i].Ok)
	}
}

func TestCalculateCheckCodex(t *testing.T) {
	code := "91350122MADDCYYE39"
	sum := 0
	for i, char := range code[:17] {
		index := strings.Index(codeOrigin, string(char))
		if index == -1 {
			t.Logf("char %s is not in codeOrigin: %s", string(char), code)
			return
		}
		weight := weightedfactors[i]
		sum += index * weight
	}
	remainder := sum % 31
	checkCode := (31 - remainder) % 31
	t.Log(checkCode)
	if checkCode > len(codeOrigin)-1 {
		t.Logf("length of codeOrigin is not enough: %s", code)
		return
	}
	assert.NotEqual(t, code[17:], string(codeOrigin[checkCode]))
}



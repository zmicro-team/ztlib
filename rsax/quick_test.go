package rsax

import (
	_ "embed"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed res/rsa-public.key
var public_key string

//go:embed res/rsa-private.key
var private_key string

var t_map = map[string]string{
	"X": "5",
	"C": "3",
	"A": "1",
	"B": "2",
	"G": "4",
}

func TestMapSortToVal(t *testing.T) {
	t.Log(MapSortToVal[string](t_map))
}

func TestPublicEncrypt(t *testing.T) {
	data := strings.Join(MapSortToVal[string](t_map), "&")
	pet, err := PublicEncrypt(data, public_key)
	assert.Equal(t, err, nil)
	ped, err := PriKeyDecrypt(pet, private_key)
	assert.Equal(t, err, nil)
	hexs, _ := hex.DecodeString(ped)
	assert.Equal(t, string(hexs), data)
}

func TestSignMd5WithRsa(t *testing.T) {
	data := strings.Join(MapSortToVal[string](t_map), "&")
	sign_data, err := SignMd5WithRsa(data, private_key)
	assert.Equal(t, err, nil)
	err = VerifySignMd5WithRsa(data, sign_data, `-----BEGIN Public key-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAk+89V7vpOj1rG6bTAKYM
56qmFLwNCBVDJ3MltVVtxVUUByqc5b6u909MmmrLBqS//PWC6zc3wZzU1+ayh8xb
UAEZuA3EjlPHIaFIVIz04RaW10+1xnby/RQE23tDqsv9a2jv/axjE/27b62nzvCW
eItu1kNQ3MGdcuqKjke+LKhQ7nWPRCOd/ffVqSuRvG0YfUEkOz/6UpsPr6vrI331
hWRB4DlYy8qFUmDsyvvExe4NjZWblXCqkEXRRAhi2SQRCl3teGuIHtDUxCskRIDi
aMD+Qt2Yp+Vvbz6hUiqIWSIH1BoHJer/JOq2/O6X3cmuppU4AdVNgy8Bq236iXvr
MQIDAQAB
-----END Public key-----
`)
	assert.NotEqual(t, err, nil)

	err = VerifySignMd5WithRsa(data, sign_data, public_key)
	assert.Equal(t, err, nil)
}

func TestKeyEncodeDecode(t *testing.T) {
	urlValues := url.Values{}
	urlValues.Add("X", "5")
	urlValues.Add("C", "3")
	data := urlValues.Encode()
	pr := New(SetPublicString(public_key),
		SetPrivateString(private_key))

	result, err := pr.pubKeyEncode([]byte(data))
	assert.Equal(t, err, nil)
	resultDecode, err := pr.priKeyDecode(result)
	assert.Equal(t, err, nil)
	assert.Equal(t, data, string(resultDecode))

	result, err = pr.priKeyEncode([]byte(data))
	assert.Equal(t, err, nil)
	resultDecode, err = pr.pubKeyDecode(result)
	assert.Equal(t, err, nil)
	assert.Equal(t, data, string(resultDecode))
}

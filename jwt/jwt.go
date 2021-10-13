package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)


func GenHS256(appKey, appSecret, uid string, expire time.Duration, extra map[string]interface{}) (string, error) {

	expTS := time.Now().Add(expire).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   appKey, // string
		"exp":   expTS,  // int64
		"uid":   uid,    // string
		"extra": extra,  // extra data
	})
	signed, err := t.SignedString([]byte(appSecret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func TestGenHS256(t *testing.T) {

	tokString, err := GenHS256("1", "1", "1", time.Minute*5, nil)
	if err != nil {
		t.Error(err)
	}
	if err != nil {
		t.Error(err)
	}
	if tokString == "" {
		t.Error(tokString)
	}

}
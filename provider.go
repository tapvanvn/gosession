package gosession

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/tapvanvn/goutil"
)

var __h = sha256.New()
var __hmd5 = md5.New()

type Provider struct {
}

func NewProvider() (*Provider, error) {

	provider := &Provider{}
	return provider, nil
}

func (pro *Provider) start() {

}

func (pro *Provider) IssueSessionString(agent interface{}) (string, error) {

	chunkID, code := getChunkCode()
	sessionID, err := incrSessionID()

	if err != nil {

		return "", err
	}
	parts := strings.Split(code, ".")

	h256 := sha256.New()

	_ = getStepSalt(chunkID, GetStep(sessionID))
	hash, err := HashStep(parts[1], chunkID, sessionID)

	if err != nil {

		return "", err
	}
	h256.Write([]byte(fmt.Sprintf("%s.%s", parts[0], hash)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))

	return fmt.Sprintf("%d.%d.%s", chunkID, sessionID, hashString), nil
}

//MARK: utility for provider

func getStepSalt(chunkID int, step int) string {
	salt, err := getSalt(chunkID, step)
	if salt == "" || err != nil {
		salt = goutil.GenSecretKey(__config.SaltLength)
		setSalt(chunkID, step, salt)
	}
	return salt
}

func getChunkCode() (int, string) {

	chunkID := CurrentChunkID()
	code, err := getSVSessionCode(chunkID)

	if code == "" || err != nil {

		code = fmt.Sprintf("%s.%s", goutil.GenSecretKey(__config.CodeLength), goutil.GenSecretKey(__config.CodeLength))
		setSVSessionCode(chunkID, code)
	}
	return chunkID, code
}

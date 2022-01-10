package gosession

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/tapvanvn/goutil"
)

func NewProviderWithQuota(quota int64) *ProviderQuota {

	return &ProviderQuota{

		Quota: quota,
	}
}

type ProviderQuota struct {
	Quota int64
}

func (pro *ProviderQuota) IssueSessionString(accessPoint string, agent string) (string, error) {

	if pro.Quota > 0 {

		current, err := getEndPoint(accessPoint)

		if err != nil {

			return "", ErrInvalidContext
		}
		if current > pro.Quota {

			return "", ErrHitEndpointQuota
		}
	}
	incrEndpoint(accessPoint)

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
	h256.Write([]byte(fmt.Sprintf("%s.%s.%s", parts[0], hash, agent)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))

	return fmt.Sprintf("%d.%d.%s", chunkID, sessionID, hashString), nil
}

func (pro *ProviderQuota) IssueRotateSessionString(accessPoint string, agent string, action int) (string, string, error) {

	if pro.Quota > 0 {

		current, err := getEndPoint(accessPoint)

		if err != nil {

			return "", "", ErrInvalidContext
		}
		if current > pro.Quota {

			return "", "", ErrHitEndpointQuota
		}
	}
	incrEndpoint(accessPoint)

	chunkID, code := getChunkCode()
	sessionID, err := incrSessionID()
	rotateCodeA := goutil.GenSecretKey(5)
	rotateCodeB := goutil.GenSecretKey(5)

	if err != nil {

		return "", "", err
	}
	parts := strings.Split(code, ".")

	h256 := sha256.New()

	_ = getStepSalt(chunkID, GetStep(sessionID))
	hash, err := HashStep(parts[1], chunkID, sessionID)

	if err != nil {

		return "", "", err
	}
	h256.Write([]byte(fmt.Sprintf("%s%s.%s.%s.%s", rotateCodeA, rotateCodeB, parts[0], hash, agent)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))

	setRotateCode(sessionID, action, rotateCodeA, time.Second*60)

	return fmt.Sprintf("%d.%d.%s", chunkID, sessionID, hashString), rotateCodeB, nil
}

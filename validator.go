package gosession

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tapvanvn/goutil"
)

func NewValidator(actionPerSecond int64) (*Validator, error) {
	validator := &Validator{
		TotalQuota:   actionPerSecond,
		ActionQuotas: map[int]int64{},
	}
	return validator, nil
}

type ActionGroup struct {
	GroupID int
	Actions []int
}

type Validator struct {
	ActionQuotas map[int]int64
	TotalQuota   int64
}

type SessionInfo struct {
	ChunkID   int
	SessionID int64
	Hash      string
}

func (val *Validator) AddActionQuota(action int, callPerSecond int64) {

	val.ActionQuotas[action] = callPerSecond
}

func (val *Validator) getInfo(sessionString string) (*SessionInfo, error) {

	parts := strings.Split(sessionString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidSession
	}
	vscodeID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	sessionID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &SessionInfo{
		ChunkID:   vscodeID,
		SessionID: sessionID,
		Hash:      parts[2],
	}, nil
}

func (val *Validator) validateSession(sessionInfo *SessionInfo, agent string) error {

	code, err := getSVSessionCode(sessionInfo.ChunkID)
	if err != nil {
		return err
	}

	if len(code) < 64 {
		return ErrInvalidSession
	}
	codeParts := strings.Split(code, ".")
	h256 := sha256.New()
	hash, err := HashStep(codeParts[1], sessionInfo.ChunkID, sessionInfo.SessionID)
	if err != nil {
		return err
	}
	h256.Write([]byte(fmt.Sprintf("%s.%s.%s", codeParts[0], hash, agent)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))
	if hashString != sessionInfo.Hash {

		return ErrInvalidSession
	}
	return nil
}

func (val *Validator) validateSessionRotate(sessionInfo *SessionInfo, agent string, action int, rotateCode string) (string, string, error) {

	code, err := getSVSessionCode(sessionInfo.ChunkID)
	if err != nil {
		fmt.Println("validateSessionRotate err:", err.Error(), sessionInfo.ChunkID)
		return "", "", err
	}
	if len(code) < 64 {
		return "", "", ErrInvalidSessionLength
	}
	rotateCodeA, err := getRotateCode(sessionInfo.SessionID, action)
	if err != nil {
		//getRotateCode err: redis: nil 1074
		fmt.Println("getRotateCode err:", err.Error(), sessionInfo.SessionID)
		return "", "", err
	}
	fmt.Println("rotateCodeA:", rotateCodeA, "rotateCodeB", rotateCode)
	codeParts := strings.Split(code, ".")
	h256 := sha256.New()
	hash, err := HashStep(codeParts[1], sessionInfo.ChunkID, sessionInfo.SessionID)
	if err != nil {
		fmt.Println("HashStep err:", err.Error(), codeParts[1], sessionInfo.ChunkID, sessionInfo.SessionID)
		return "", "", err
	}
	h256.Write([]byte(fmt.Sprintf("%s%s.%s.%s.%s", rotateCodeA, rotateCode, codeParts[0], hash, agent)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))

	if hashString != sessionInfo.Hash {

		fmt.Println("should:", hashString, sessionInfo.Hash)
		return "", "", ErrInvalidSessionVerifyHashFail
	}
	chunkID, code := getChunkCode()
	rotateCodeA = goutil.GenSecretKey(5)
	rotateCodeB := goutil.GenSecretKey(5)
	parts := strings.Split(code, ".")
	h256 = sha256.New()

	_, err = getStepSalt(chunkID, GetStep(sessionInfo.SessionID))
	if err != nil {
		fmt.Println("getStepSalt err:", err.Error())
		return "", "", err
	}
	hash, err = HashStep(parts[1], chunkID, sessionInfo.SessionID)

	if err != nil {
		fmt.Println("HashStep err:", err.Error())
		return "", "", err
	}

	fmt.Println("new rotateCodeA:", rotateCodeA, "rotateCodeB", rotateCodeB)

	h256.Write([]byte(fmt.Sprintf("%s%s.%s.%s.%s", rotateCodeA, rotateCodeB, parts[0], hash, agent)))
	hmd5 = md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString = hex.EncodeToString(hmd5.Sum(nil))

	setRotateCode(sessionInfo.SessionID, action, rotateCodeA, time.Second*180)

	return fmt.Sprintf("%d.%d.%s", chunkID, sessionInfo.SessionID, hashString), rotateCodeB, nil
}

func (val *Validator) Validate(sessionString string, agent string) error {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {

		return ErrInvalidSession
	}
	return val.validateSession(sessionInfo, agent)
}

func (val *Validator) ValidateRotate(sessionString string, agent string, rotateCode string) (string, string, error) {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {

		return "", "", ErrInvalidSession
	}
	return val.validateSessionRotate(sessionInfo, agent, 0, rotateCode)
}

func (val *Validator) ValidateAction(sessionString string, agent string, action int) error {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {
		//fmt.Println("cannot get sessionInfo")
		return ErrInvalidSession
	}

	if err := val.validateSession(sessionInfo, agent); err != nil {
		//fmt.Println("err:", err.Error())
		return ErrInvalidSession
	}
	if val.TotalQuota > 0 {
		if quota, err := getTotalQuota(sessionInfo.SessionID); err == nil {
			if quota > val.TotalQuota {
				return ErrHitQuota
			}
			incrTotalQuota(sessionInfo.SessionID)
		}
	}
	if actionQuota, ok := val.ActionQuotas[action]; ok && actionQuota > 0 {

		if quota, err := getActionQuota(sessionInfo.SessionID, action); err == nil {
			//fmt.Println("quota", sessionString, action, quota)
			if quota > actionQuota {
				return ErrHitQuota
			}
			if quota == 0 {
				setActionQuota(sessionInfo.SessionID, action)
			} else {
				incrActionQuota(sessionInfo.SessionID, action)
			}
		}
	}
	return nil
}

func (val *Validator) ValidateRotateAction(sessionString string, agent string, action int, rotateCode string) (string, string, error) {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {
		//fmt.Println("cannot get sessionInfo")
		return "", "", err
	}

	newSessionString, newRotateCode, err := val.validateSessionRotate(sessionInfo, agent, action, rotateCode)
	if err != nil {
		//fmt.Println("err:", err.Error())
		return "", "", err
	}
	if val.TotalQuota > 0 {
		if quota, err := getTotalQuota(sessionInfo.SessionID); err == nil {
			if quota > val.TotalQuota {
				return "", "", ErrHitQuota
			}
			incrTotalQuota(sessionInfo.SessionID)
		}
	}
	if actionQuota, ok := val.ActionQuotas[action]; ok && actionQuota > 0 {

		if quota, err := getActionQuota(sessionInfo.SessionID, action); err == nil {
			//fmt.Println("quota", sessionString, action, quota)
			if quota > actionQuota {
				return "", "", ErrHitQuota
			}
			if quota == 0 {
				setActionQuota(sessionInfo.SessionID, action)
			} else {
				incrActionQuota(sessionInfo.SessionID, action)
			}
		}
	}
	return newSessionString, newRotateCode, nil
}

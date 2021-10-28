package gosession

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
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

func (val *Validator) validateSession(sessionInfo *SessionInfo) error {

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
	h256.Write([]byte(fmt.Sprintf("%s.%s", codeParts[0], hash)))
	hmd5 := md5.New()
	hmd5.Write(h256.Sum(nil))

	hashString := hex.EncodeToString(hmd5.Sum(nil))
	if hashString != sessionInfo.Hash {

		return ErrInvalidSession
	}
	return nil
}

func (val *Validator) Validate(sessionString string) error {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {

		return ErrInvalidSession
	}
	return val.validateSession(sessionInfo)
}

func (val *Validator) ValidateAction(sessionString string, action int) error {

	sessionInfo, err := val.getInfo(sessionString)
	if err != nil {
		return ErrInvalidSession
	}

	if err := val.validateSession(sessionInfo); err != nil {
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
			fmt.Println("quota", sessionString, action, quota)
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

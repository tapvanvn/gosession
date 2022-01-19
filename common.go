package gosession

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tapvanvn/godbengine/engine"
)

const (
	KEY_SESSION_ID     = "sid"
	KEY_SVSESSION_CODE = "svcode"
	KEY_STEP_SALT      = "ssalt"
	KEY_ACCESS_POINT   = "spoint"
	KEY_TOTAL_QUOTA    = "sqtotal"
	KEY_ACTION_QUOTA   = "sqact"
	KEY_ROTATE_CODE    = "srcode"
)

var (
	//MARK: Key define
	KeySessionID     = KEY_SESSION_ID
	KeySVSessionCode = KEY_SVSESSION_CODE
	KeyStepSalt      = KEY_STEP_SALT
	KeyAccessPoint   = KEY_ACCESS_POINT
	KeyTotalQuota    = KEY_TOTAL_QUOTA
	KeyActionQuota   = KEY_ACTION_QUOTA
	KeyRotateCode    = KEY_ROTATE_CODE

	//MARK:
	__config *Config        = nil
	__eng    *engine.Engine = nil
)

var ErrInvalidSession = errors.New("Invalid Session")
var ErrInvalidSessionVerifyHashFail = errors.New("Invalid Session Hash Fail")
var ErrInvalidSessionLength = errors.New("Invalid Session Length")
var ErrInvalidConfig = errors.New("Invalid Config")
var ErrInvalidContext = errors.New("Invalid Context")
var ErrHitEndpointQuota = errors.New("Hit Endpoint Quota")
var ErrHitQuota = errors.New("Hit Quota")

func GetStep(sessionID int64) int {

	return int(sessionID % int64(__config.StepNum))
}
func HashStep(baseString string, chunkID int, sessionID int64) (string, error) {

	step := GetStep(sessionID)
	salt, err := getSalt(chunkID, step)
	if err != nil {
		return "", err
	}
	calcString := baseString

	hmd5 := md5.New()

	hmd5.Write([]byte(salt))

	for i := __config.StepMin; i < __config.StepMin+step; i++ {

		hmd5.Write([]byte(calcString))
	}

	hmd5.Write([]byte(salt))

	hash := hex.EncodeToString(hmd5.Sum(nil))

	return hash, nil
}

func GetKeySVSessionCode(chunkID int) string {

	return fmt.Sprintf("%s_%d", KeySVSessionCode, chunkID)
}

func GetKeyStepSalt(chunkID int, step int) string {

	return fmt.Sprintf("%s_%d_%d", KeyStepSalt, chunkID, step)
}

func GetKeyAccessPoint(accessPoint string) string {
	daySecond := time.Now().Unix() % 86400
	return fmt.Sprintf("%s_%d_%s", KeyAccessPoint, daySecond, accessPoint)
}

func GetKeyTotalQuota(sessionID int64) string {

	daySecond := time.Now().Unix() % 86400
	return fmt.Sprintf("%s_%d_%d", KeyTotalQuota, daySecond, sessionID)
}
func GetKeyActionQuota(sessionID int64, action int) string {

	daySecond := time.Now().Unix() % 86400
	return fmt.Sprintf("%s_%d_%d_%d", KeyTotalQuota, daySecond, sessionID, action)
}
func GetKeyRotateCode(sessionID int64, action int) string {

	return fmt.Sprintf("%s_%d_%d", KeyRotateCode, sessionID, action)
}

func Init(config *Config, eng *engine.Engine) error {

	if eng == nil {
		return ErrInvalidConfig
	}
	if config == nil {
		__config = DefaultConfig
	} else {
		__config = config
	}
	if len(strings.TrimSpace(__config.KeyPrefix)) > 0 {

		KeySessionID = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_SESSION_ID)
		KeySVSessionCode = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_SVSESSION_CODE)
		KeyAccessPoint = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_ACCESS_POINT)
		KeyTotalQuota = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_TOTAL_QUOTA)
		KeyActionQuota = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_ACTION_QUOTA)
		KeyRotateCode = fmt.Sprintf("%s_%s", __config.KeyPrefix, KEY_ROTATE_CODE)
	}

	if __config.ChunkDuration <= 0 {

		__config.ChunkDuration = 10
	}

	if __config.StepMin < 10 {

		__config.StepMin = 10
	}

	if __config.StepNum < 10 {

		__config.StepNum = 10
	}
	__eng = eng
	return nil
}

func CurrentChunkID() int {

	dayTime := int(time.Now().Unix() % 86400)
	return dayTime / __config.ChunkDuration
}

type IProvider interface {
	IssueSessionString(endpoint string, agent string) (string, error)
	IssueRotateSessionString(endpoint string, agent string, action int) (string, string, error)
}

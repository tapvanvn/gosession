package gosession

import "time"

func getSessionID() (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.GetInt(KeySessionID)
}

func incrSessionID() (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.IncrInt(KeySessionID)
}

//MARK: Server Session Code
func setSVSessionCode(chunkID int, code string) error {
	memPool := __eng.GetMemPool()
	return memPool.Set(GetKeySVSessionCode(chunkID), code)
}

func getSVSessionCode(chunkID int) (string, error) {
	memPool := __eng.GetMemPool()
	return memPool.Get(GetKeySVSessionCode(chunkID))
}

//MARK: salt
func setSalt(chunkID int, step int, salt string) error {
	memPool := __eng.GetMemPool()
	return memPool.Set(GetKeyStepSalt(chunkID, step), salt)
}

func getSalt(chunkID int, step int) (string, error) {
	memPool := __eng.GetMemPool()
	return memPool.Get(GetKeyStepSalt(chunkID, step))
}

//MARK: endpoint - 1 key per agent
func incrEndpoint(endpoint string) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.IncrInt(GetKeyAccessPoint(endpoint))
}

func getEndPoint(endpoint string) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.GetInt(GetKeyAccessPoint(endpoint))
}

//MARK: Total Quota - 1 key per session
func incrTotalQuota(sessionID int64) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.IncrInt(GetKeyTotalQuota(sessionID))
}

func getTotalQuota(sessionID int64) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.GetInt(GetKeyTotalQuota(sessionID))
}

func setTotalQuota(sessionID int64) (int64, error) {
	memPool := __eng.GetMemPool()
	err := memPool.SetIntExpire(GetKeyTotalQuota(sessionID), 1, time.Second)
	return 1, err
}

//MARK: Action Quota - 1 key per (action + session)
func incrActionQuota(sessionID int64, action int) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.IncrInt(GetKeyActionQuota(sessionID, action))
}

func getActionQuota(sessionID int64, action int) (int64, error) {
	memPool := __eng.GetMemPool()
	return memPool.GetInt(GetKeyActionQuota(sessionID, action))
}

func setActionQuota(sessionID int64, action int) (int64, error) {
	memPool := __eng.GetMemPool()
	err := memPool.SetIntExpire(GetKeyActionQuota(sessionID, action), 1, time.Second)
	return 1, err
}

//MARK: Rotate //1 key - per session
func getRotateCode(sessionID int64, action int) (string, error) {
	memPool := __eng.GetMemPool()
	return memPool.Get(GetKeyRotateCode(sessionID, action))
}

func setRotateCode(sessionID int64, action int, code string, duration time.Duration) error {

	memPool := __eng.GetMemPool()
	return memPool.SetExpire(GetKeyRotateCode(sessionID, action), code, duration)
}

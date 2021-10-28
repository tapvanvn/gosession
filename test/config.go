package test

type MemDBConfig struct {
	ConnectionString string `json:"ConnectionString"`
}

type DocumentDBConfig struct {
	Provider         string `json:"Provider"`
	ConnectionString string `json:"ConnectionString"`
	DatabaseName     string `json:"DatabaseName"`
}

//SystemConfig system config
type Config struct {
	MemDB      *MemDBConfig      `json:"MemDB"`
	DocumentDB *DocumentDBConfig `json:"DocumentDB"`
}

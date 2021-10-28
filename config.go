package gosession

type Config struct {
	KeyPrefix     string
	ChunkDuration int //in second
	StepMin       int //mininum hash step
	StepNum       int //num of step
	CodeLength    int //the length of session code
	SaltLength    int //the length of step salt
}

var DefaultConfig = &Config{
	KeyPrefix:     "",
	ChunkDuration: 10,
	StepMin:       50,
	StepNum:       100,
	CodeLength:    32,
	SaltLength:    5,
}

package storage

type (
	// DBConfiguration describes configuration for DB. Supported: 'levelDB', 'redisDB', 'boltDB', 'badgerDB'.
	DBConfiguration struct {
		Type           string         `yaml:"Type"`
		LevelDBOptions LevelDBOptions `yaml:"LevelDBOptions"`
	}
)

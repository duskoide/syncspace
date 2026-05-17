package config

import "os"

type Config struct {
	Addr   string
	DBPath string
}

func Load() Config {
	addr := os.Getenv("SYNCSPACE_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dbPath := os.Getenv("SYNCSPACE_DB_PATH")
	if dbPath == "" {
		dbPath = "../data/syncspace.db"
	}
	return Config{Addr: addr, DBPath: dbPath}
}

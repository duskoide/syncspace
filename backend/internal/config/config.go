package config

import "os"

type Config struct {
	Addr      string
	DBPath    string
	UploadDir string
	JWTSecret string
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
	uploadDir := os.Getenv("SYNCSPACE_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	jwtSecret := os.Getenv("SYNCSPACE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "syncspace-edu-secret-key-change-in-production"
	}
	return Config{Addr: addr, DBPath: dbPath, UploadDir: uploadDir, JWTSecret: jwtSecret}
}

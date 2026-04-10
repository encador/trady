package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("[ERROR] No DB File Specified")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("[ERROR] DB File (%s) Does NOT Exists", path)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func Create(path string) error {
	if path == "" {
		return errors.New("[ERROR] No DB File Name Specified")
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("[ERROR] DB File (%s) Already Exists", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, nil, 0o644)
}

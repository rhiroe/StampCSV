package config

import (
	"os"
	"path/filepath"
)

const appName = "StampCSV"

func configFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, "dir.txt"), nil
}

// LoadDir は保存済みのディレクトリパスを返す。未設定の場合は空文字を返す。
func LoadDir() string {
	path, err := configFilePath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveDir はディレクトリパスを設定ファイルに保存する。
func SaveDir(dir string) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(dir), 0o600)
}

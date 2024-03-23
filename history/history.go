package history

import (
	"os"
	"path/filepath"
)

type Store struct {
	filepath string
}

func NewStore(fp string) (*Store, error) {
	if fp == "" {
		path, err := defaultFilepath()
		if err != nil {
			return nil, err
		}
		fp = path
	}
	if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
		return nil, err
	}
	return &Store{filepath: fp}, nil
}

func (s *Store) Append(command string) error {
	f, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(command + "\n"); err != nil {
		return err
	}

	// TODO: Limit the number of lines in the file.
	return nil
}

func defaultFilepath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, ".kubehelp", "history.txt"), nil
}

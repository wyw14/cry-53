package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct{ root string }

func NewLocalStore(root string) (*LocalStore, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &LocalStore{root: abs}, nil
}

func (s *LocalStore) Save(ctx context.Context, name string, reader io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	extension := strings.ToLower(filepath.Ext(name))
	if extension != ".json" {
		return "", fmt.Errorf("unsupported attachment type")
	}
	temporary, err := os.CreateTemp(s.root, "upload-*.tmp")
	if err != nil {
		return "", err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(temporary, hash), reader); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	id := hex.EncodeToString(hash.Sum(nil)) + extension
	target := filepath.Join(s.root, id)
	if err := os.Rename(temporaryName, target); err != nil {
		if _, statErr := os.Stat(target); statErr == nil {
			return id, nil
		}
		return "", err
	}
	return id, nil
}

func (s *LocalStore) Open(ctx context.Context, id string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if filepath.Base(id) != id || strings.Contains(id, "..") {
		return nil, fmt.Errorf("invalid attachment id")
	}
	return os.Open(filepath.Join(s.root, id))
}

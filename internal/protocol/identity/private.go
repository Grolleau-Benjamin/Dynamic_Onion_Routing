package identity

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"golang.org/x/crypto/curve25519"
)

type PrivateIdentity struct {
	UUID    [16]byte
	PrivKey [32]byte
	PubKey  [32]byte
}

type identityStore struct {
	dir      string
	uuidPath string
	privPath string
	pubPath  string
}

func newIdentityStore(dir string) identityStore {
	return identityStore{
		dir:      dir,
		uuidPath: filepath.Join(dir, "relay.uuid"),
		privPath: filepath.Join(dir, "relay.priv"),
		pubPath:  filepath.Join(dir, "relay.pub"),
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func loadUUID(path string) ([16]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return [16]byte{}, err
	}

	id, err := uuid.Parse(string(raw))
	if err != nil {
		return [16]byte{}, errors.New("invalid UUID format")
	}

	return id, nil
}

func generateUUID(path string) ([16]byte, error) {
	id := uuid.New()
	if err := os.WriteFile(path, []byte(id.String()), 0644); err != nil {
		return [16]byte{}, err
	}
	return id, nil
}

func loadKey32(path string) ([32]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	if len(raw) != 32 {
		return [32]byte{}, fmt.Errorf("invalid key size: expected 32 bytes")
	}

	var key [32]byte
	copy(key[:], raw)
	return key, nil
}

func generatePrivKey(path string) ([32]byte, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return [32]byte{}, err
	}

	if err := os.WriteFile(path, priv[:], 0600); err != nil {
		return [32]byte{}, err
	}
	return priv, nil
}

func generatePubKey(path string, priv [32]byte) ([32]byte, error) {
	pubSlice, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return [32]byte{}, err
	}

	var pub [32]byte
	copy(pub[:], pubSlice)

	if err := os.WriteFile(path, pub[:], 0644); err != nil {
		return [32]byte{}, err
	}
	return pub, nil
}

func LoadPrivateIdentity(dir string) (*PrivateIdentity, error) {
	store := newIdentityStore(dir)
	pi := &PrivateIdentity{}

	if err := os.MkdirAll(store.dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create identity dir: %w", err)
	}

	var err error

	if fileExists(store.uuidPath) {
		pi.UUID, err = loadUUID(store.uuidPath)
	} else {
		pi.UUID, err = generateUUID(store.uuidPath)
	}
	if err != nil {
		return nil, fmt.Errorf("UUID error: %w", err)
	}

	if fileExists(store.privPath) {
		pi.PrivKey, err = loadKey32(store.privPath)
	} else {
		pi.PrivKey, err = generatePrivKey(store.privPath)
	}
	if err != nil {
		return nil, fmt.Errorf("private key error: %w", err)
	}

	if fileExists(store.pubPath) {
		pi.PubKey, err = loadKey32(store.pubPath)
	} else {
		pi.PubKey, err = generatePubKey(store.pubPath, pi.PrivKey)
	}
	if err != nil {
		return nil, err
	}

	return pi, nil
}

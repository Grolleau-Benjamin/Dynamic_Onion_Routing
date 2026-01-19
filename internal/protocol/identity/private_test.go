package identity_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/internal/protocol/identity"
	"github.com/google/uuid"
	"golang.org/x/crypto/curve25519"
)

func TestLoadPrivateIdentity_NewIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pi, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pi == nil {
		t.Fatal("expected non-nil PrivateIdentity")
	}

	if pi.UUID == [16]byte{} {
		t.Fatal("UUID should not be empty")
	}

	if pi.PrivKey == [32]byte{} {
		t.Fatal("PrivKey should not be empty")
	}

	if pi.PubKey == [32]byte{} {
		t.Fatal("PubKey should not be empty")
	}

	expectedPub, err := curve25519.X25519(pi.PrivKey[:], curve25519.Basepoint)
	if err != nil {
		t.Fatalf("failed to derive public key: %v", err)
	}

	var expectedPubArray [32]byte
	copy(expectedPubArray[:], expectedPub)

	if pi.PubKey != expectedPubArray {
		t.Fatal("public key does not match derived key from private key")
	}

	uuidPath := filepath.Join(dir, "relay.uuid")
	privPath := filepath.Join(dir, "relay.priv")
	pubPath := filepath.Join(dir, "relay.pub")

	if _, err := os.Stat(uuidPath); os.IsNotExist(err) {
		t.Fatal("UUID file was not created")
	}

	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		t.Fatal("private key file was not created")
	}

	if _, err := os.Stat(pubPath); os.IsNotExist(err) {
		t.Fatal("public key file was not created")
	}
}

func TestLoadPrivateIdentity_LoadExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pi1, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	pi2, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}

	if pi1.UUID != pi2.UUID {
		t.Fatalf("UUID mismatch:\n\tfirst:  %v\n\tsecond: %v", pi1.UUID, pi2.UUID)
	}

	if pi1.PrivKey != pi2.PrivKey {
		t.Fatal("PrivKey should be the same across loads")
	}

	if pi1.PubKey != pi2.PubKey {
		t.Fatal("PubKey should be the same across loads")
	}
}

func TestLoadPrivateIdentity_InvalidUUID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	uuidPath := filepath.Join(dir, "relay.uuid")

	err := os.WriteFile(uuidPath, []byte("invalid-uuid"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid UUID: %v", err)
	}

	_, err = identity.LoadPrivateIdentity(dir)
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestLoadPrivateIdentity_InvalidPrivKeySize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	privPath := filepath.Join(dir, "relay.priv")

	err := os.WriteFile(privPath, []byte("short"), 0600)
	if err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	_, err = identity.LoadPrivateIdentity(dir)
	if err == nil {
		t.Fatal("expected error for invalid private key size")
	}
}

func TestLoadPrivateIdentity_InvalidPubKeySize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	validUUID := uuid.New()
	uuidPath := filepath.Join(dir, "relay.uuid")
	err := os.WriteFile(uuidPath, []byte(validUUID.String()), 0644)
	if err != nil {
		t.Fatalf("failed to write UUID: %v", err)
	}

	var validPriv [32]byte
	copy(validPriv[:], []byte("this_is_a_32_byte_private_key!!"))
	privPath := filepath.Join(dir, "relay.priv")
	err = os.WriteFile(privPath, validPriv[:], 0600)
	if err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	pubPath := filepath.Join(dir, "relay.pub")
	err = os.WriteFile(pubPath, []byte("short"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid public key: %v", err)
	}

	_, err = identity.LoadPrivateIdentity(dir)
	if err == nil {
		t.Fatal("expected error for invalid public key size")
	}
}

func TestLoadPrivateIdentity_PubKeyMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	validUUID := uuid.New()
	uuidPath := filepath.Join(dir, "relay.uuid")
	err := os.WriteFile(uuidPath, []byte(validUUID.String()), 0644)
	if err != nil {
		t.Fatalf("failed to write UUID: %v", err)
	}

	var validPriv [32]byte
	for i := range validPriv {
		validPriv[i] = byte(i)
	}
	validPriv[0] &= 248
	validPriv[31] &= 127
	validPriv[31] |= 64

	privPath := filepath.Join(dir, "relay.priv")
	err = os.WriteFile(privPath, validPriv[:], 0600)
	if err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	var wrongPub [32]byte
	for i := range wrongPub {
		wrongPub[i] = byte(255 - i)
	}

	pubPath := filepath.Join(dir, "relay.pub")
	err = os.WriteFile(pubPath, wrongPub[:], 0644)
	if err != nil {
		t.Fatalf("failed to write wrong public key: %v", err)
	}

	pi, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPub, err := curve25519.X25519(validPriv[:], curve25519.Basepoint)
	if err != nil {
		t.Fatalf("failed to derive expected public key: %v", err)
	}

	var expectedPubArray [32]byte
	copy(expectedPubArray[:], expectedPub)

	if pi.PubKey != expectedPubArray {
		t.Fatal("public key should be repaired to match private key")
	}

	loadedPub, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatalf("failed to read repaired public key: %v", err)
	}

	var loadedPubArray [32]byte
	copy(loadedPubArray[:], loadedPub)

	if loadedPubArray != expectedPubArray {
		t.Fatal("public key file should be overwritten with correct value")
	}
}

func TestLoadPrivateIdentity_DirectoryCreation(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "nested", "identity", "dir")

	_, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal("directory was not created")
	}

	if !info.IsDir() {
		t.Fatal("path is not a directory")
	}
}

func TestLoadPrivateIdentity_FilePermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	privPath := filepath.Join(dir, "relay.priv")
	info, err := os.Stat(privPath)
	if err != nil {
		t.Fatalf("failed to stat private key file: %v", err)
	}

	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Fatalf("private key has wrong permissions: got %o, want 0600", mode.Perm())
	}

	pubPath := filepath.Join(dir, "relay.pub")
	info, err = os.Stat(pubPath)
	if err != nil {
		t.Fatalf("failed to stat public key file: %v", err)
	}

	mode = info.Mode()
	if mode.Perm() != 0644 {
		t.Fatalf("public key has wrong permissions: got %o, want 0644", mode.Perm())
	}
}

func TestLoadPrivateIdentity_UUIDPersistence(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pi1, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	uuidPath := filepath.Join(dir, "relay.uuid")
	uuidBytes, err := os.ReadFile(uuidPath)
	if err != nil {
		t.Fatalf("failed to read UUID file: %v", err)
	}

	parsedUUID, err := uuid.Parse(string(uuidBytes))
	if err != nil {
		t.Fatalf("UUID file contains invalid UUID: %v", err)
	}

	if parsedUUID != pi1.UUID {
		t.Fatal("UUID in file does not match loaded UUID")
	}
}

func TestLoadPrivateIdentity_PrivKeyFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pi, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pi.PrivKey[0]&7 != 0 {
		t.Fatal("private key should have lowest 3 bits of first byte cleared")
	}

	if pi.PrivKey[31]&128 != 0 {
		t.Fatal("private key should have highest bit of last byte cleared")
	}

	if pi.PrivKey[31]&64 == 0 {
		t.Fatal("private key should have second highest bit of last byte set")
	}
}

func TestLoadPrivateIdentity_ConcurrentLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pi1, err := identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			pi, err := identity.LoadPrivateIdentity(dir)
			if err != nil {
				t.Errorf("concurrent load failed: %v", err)
			}
			if pi.UUID != pi1.UUID {
				t.Errorf("UUID mismatch in concurrent load")
			}
			if pi.PrivKey != pi1.PrivKey {
				t.Errorf("PrivKey mismatch in concurrent load")
			}
			if pi.PubKey != pi1.PubKey {
				t.Errorf("PubKey mismatch in concurrent load")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoadPrivateIdentity_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(entries) != 0 {
		t.Fatal("directory should be empty initially")
	}

	_, err = identity.LoadPrivateIdentity(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err = os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 files, got %d", len(entries))
	}

	expectedFiles := map[string]bool{
		"relay.uuid": false,
		"relay.priv": false,
		"relay.pub":  false,
	}

	for _, entry := range entries {
		if _, ok := expectedFiles[entry.Name()]; ok {
			expectedFiles[entry.Name()] = true
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Fatalf("expected file %s was not created", name)
		}
	}
}

func TestLoadPrivateIdentity_ReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "readonly")

	err := os.Mkdir(dir, 0500)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	_, err = identity.LoadPrivateIdentity(dir)
	if err == nil {
		t.Fatal("expected error when writing to read-only directory")
	}
}

func TestLoadPrivateIdentity_PartialFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(string) error
	}{
		{
			name: "only UUID exists",
			setupFunc: func(dir string) error {
				uuidPath := filepath.Join(dir, "relay.uuid")
				id := uuid.New()
				return os.WriteFile(uuidPath, []byte(id.String()), 0644)
			},
		},
		{
			name: "UUID and privkey exist",
			setupFunc: func(dir string) error {
				uuidPath := filepath.Join(dir, "relay.uuid")
				id := uuid.New()
				if err := os.WriteFile(uuidPath, []byte(id.String()), 0644); err != nil {
					return err
				}

				var priv [32]byte
				for i := range priv {
					priv[i] = byte(i)
				}
				priv[0] &= 248
				priv[31] &= 127
				priv[31] |= 64

				privPath := filepath.Join(dir, "relay.priv")
				return os.WriteFile(privPath, priv[:], 0600)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			if err := tt.setupFunc(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			pi, err := identity.LoadPrivateIdentity(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pi == nil {
				t.Fatal("expected non-nil PrivateIdentity")
			}

			if pi.UUID == [16]byte{} {
				t.Fatal("UUID should not be empty")
			}

			if pi.PrivKey == [32]byte{} {
				t.Fatal("PrivKey should not be empty")
			}

			if pi.PubKey == [32]byte{} {
				t.Fatal("PubKey should not be empty")
			}
		})
	}
}

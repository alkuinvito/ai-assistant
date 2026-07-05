package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

type Encrypted string

// Value implements the driver.Valuer interface (Encrypts before saving to DB)
func (es Encrypted) Value() (driver.Value, error) {
	encryptionKey, err := hex.DecodeString(os.Getenv("DATABASE_ENCRYPTION_KEY"))
	if err != nil || encryptionKey == nil {
		return nil, errors.New("invalid encryption key")
	}

	if len(es) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the ciphertext to the nonce, keeping them together
	ciphertext := gcm.Seal(nonce, nonce, []byte(es), nil)
	return ciphertext, nil
}

// Scan implements the sql.Scanner interface (Decrypts when reading from DB)
func (es *Encrypted) Scan(value interface{}) error {
	encryptionKey, err := hex.DecodeString(os.Getenv("DATABASE_ENCRYPTION_KEY"))
	if err != nil || encryptionKey == nil {
		return errors.New("invalid encryption key")
	}

	if value == nil {
		*es = ""
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal EncryptedString value: %v", value)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(bytes) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := bytes[:nonceSize], bytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	*es = Encrypted(plaintext)
	return nil
}

package lanky_crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

// LankyCrypto is an interface that defines the methods for performing cryptographic operations.
type LankyCrypto interface {
	// ToBytes converts the given data to a byte slice.
	// It returns the byte slice representation of the data and an error if any occurred.
	ToBytes(data any) ([]byte, error)

	// Encrypt encrypts the given byte slice and returns the encryption as a string.
	// It returns the encryption string and an error if any occurred.
	Encrypt(data []byte) (encryption string, err error)

	// EncryptToBytes encrypts the given byte slice and returns the encryption as a byte slice.
	// It returns the encryption byte slice and an error if any occurred.
	EncryptToBytes(data []byte) (encryption []byte, err error)

	// Decrypt decrypts the given encryption string and returns the decrypted byte slice.
	// It returns the decrypted byte slice and an error if any occurred.
	Decrypt(encryption string) (result []byte, err error)

	// DecryptFromBytes decrypts the given encryption byte slice and returns the decrypted byte slice.
	// It returns the decrypted byte slice and an error if any occurred.
	DecryptFromBytes(encryption []byte) (result []byte, err error)
}

type lc struct {
	secret string
	size   []byte
}

// NewLankyCrypto creates a new instance of LankyCrypto with the given secret.
// It generates a random 16-byte block and initializes the LankyCrypto instance
// with the secret and the generated block.
//
// Parameters:
//   - secret: The secret used for encryption.
//
// Returns:
//   - LankyCrypto: A new instance of LankyCrypto.
func NewLankyCrypto(secret string) LankyCrypto {
	blockBytes := make([]byte, 16)
	rand.Read(blockBytes)

	return &lc{secret: secret, size: blockBytes}
}

func (c *lc) ToBytes(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (c *lc) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher([]byte(c.secret))
	if err != nil {
		return "", err
	}

	plainText := data
	cfb := cipher.NewCFBEncrypter(block, c.size)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)

	return c.encode(cipherText), nil
}

func (c *lc) EncryptToBytes(data []byte) ([]byte, error) {
	enc, err := c.Encrypt(data)
	if err != nil {
		return nil, err
	}
	return []byte(enc), nil
}

func (c *lc) Decrypt(encryption string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(c.secret))
	if err != nil {
		return nil, err
	}

	cipherText, err := c.decode(encryption)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBDecrypter(block, c.size)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

func (c *lc) DecryptFromBytes(encryption []byte) ([]byte, error) {
	dcr, err := c.Decrypt(string(encryption))
	if err != nil {
		return nil, err
	}
	return dcr, nil
}

// encode encodes the given byte slice using base64 encoding and returns the encoded string.
// It takes a byte slice as input and returns a string.
func (c *lc) encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// decode decodes a base64 encoded string and returns the decoded byte slice.
// It takes a string as input and returns a byte slice and an error.
// If the decoding is successful, the error will be nil.
// If an error occurs during decoding, the error will be non-nil.s
func (c *lc) decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

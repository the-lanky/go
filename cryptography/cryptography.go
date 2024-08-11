package lanky_crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

type LankyCrypto interface {
	ToBytes(data any) ([]byte, error)
	Encrypt(data []byte) (encryption string, err error)
	EncryptToBytes(data []byte) (encryption []byte, err error)
	Decrypt(encryption string) (result []byte, err error)
	DecryptFromBytes(encryption []byte) (result []byte, err error)
}

type lc struct {
	secret string
	size   []byte
}

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

func (c *lc) encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func (c *lc) decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

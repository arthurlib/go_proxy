package ciphers

import (
	"encoding/base64"
	"math/rand"
)

var (
	PasswordLength = 256
)

func RandomPassword() []byte {
	key := make([]byte, PasswordLength)
	for i := 0; i < PasswordLength; i++ {
		key[i] = byte(i)
	}
	rand.Shuffle(len(key), func(i, j int) {
		key[i], key[j] = key[j], key[i]
	})
	return key
}

func DumpsPassword(key []byte) string {
	return base64.RawStdEncoding.EncodeToString(key)
}

func LoadsPassword(key string) []byte {
	decoded, _ := base64.RawStdEncoding.DecodeString(key)
	return decoded
}

type CaesarCipher struct {
	encryptedPasswd []byte
	decryptPasswd   []byte
}

func (c *CaesarCipher) Encrypt(bs []byte) []byte {
	data := make([]byte, len(bs))
	for i, v := range bs {
		data[i] = c.encryptedPasswd[v]
	}
	return data
}

func (c *CaesarCipher) Decrypt(bs []byte) []byte {
	data := make([]byte, len(bs))
	for i, v := range bs {
		data[i] = c.decryptPasswd[v]
	}
	return data
}

// NewCaesarCipher 必须new一个，不可以共用
func NewCaesarCipher(key string) *CaesarCipher {
	encryptedPasswd := LoadsPassword(key)
	decryptPasswd := make([]byte, PasswordLength)
	for i, v := range encryptedPasswd {
		decryptPasswd[v] = byte(i)
	}
	return &CaesarCipher{
		encryptedPasswd,
		decryptPasswd,
	}
}

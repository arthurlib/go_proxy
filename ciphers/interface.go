package ciphers

type Cipher interface {
	Encrypt(bs []byte) []byte
	Decrypt(bs []byte) []byte
}

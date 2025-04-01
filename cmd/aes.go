package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var seed string = "This#2023$&@Spring[Sch00l)"
var key string = SHA256String(seed)[0:24]
var iv string = SHA256String(seed)[4:20]

func Encrypt(plainText string) string {
	data, err := aesCBCEncrypt([]byte(plainText), []byte(key), []byte(iv))
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(data)
}

func Decrypt(cipherText string) string {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return ""
	}

	dnData, err := aesCBCDecrypt(data, []byte(key), []byte(iv))
	if err != nil {
		return ""
	}

	return string(dnData)

}

func aesCBCEncrypt(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	plaintext = paddingPKCS7(plaintext, aes.BlockSize)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(plaintext, plaintext)

	return plaintext, nil
}

func aesCBCDecrypt(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	result := unPaddingPKCS7(ciphertext)
	return result, nil
}

func paddingPKCS7(plaintext []byte, blockSize int) []byte {
	paddingSize := blockSize - len(plaintext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
	return append(plaintext, paddingText...)
}

func unPaddingPKCS7(s []byte) []byte {
	length := len(s)
	if length == 0 {
		return s
	}
	unPadding := int(s[length-1])
	//DebugWarn("unPaddingPKCS7", length, ":", unPadding)
	if length >= unPadding {
		return s[:(length - unPadding)]
	}
	return nil
}

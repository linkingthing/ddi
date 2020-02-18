package encrypt

import (
	"bytes"

	"crypto/aes"

	"crypto/cipher"

	"encoding/base64"

	"encoding/hex"

	"errors"

	"fmt"
)

//填充

func pad(src []byte) []byte {

	padding := aes.BlockSize - len(src)%aes.BlockSize

	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(src, padtext...)

}

func unpad(src []byte) ([]byte, error) {

	length := len(src)

	unpadding := int(src[length-1])

	if unpadding > length {

		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")

	}

	return src[:(length - unpadding)], nil

}

func Encrypt(key []byte, text string) (string, error) {

	block, err := aes.NewCipher(key)

	if err != nil {

		return "", err

	}

	msg := pad([]byte(text))

	ciphertext := make([]byte, aes.BlockSize+len(msg))

	//没有向量，用的空切片

	iv := make([]byte, aes.BlockSize)

	mode := cipher.NewCBCEncrypter(block, iv)

	mode.CryptBlocks(ciphertext[aes.BlockSize:], msg)

	//finalMsg := (base64.StdEncoding.EncodeToString(ciphertext))
	finalMsg := hex.EncodeToString([]byte(ciphertext[aes.BlockSize:]))

	fmt.Println(hex.EncodeToString([]byte(ciphertext[aes.BlockSize:])))

	return finalMsg, nil

}

func Decrypt(key []byte, text string) (string, error) {

	block, err := aes.NewCipher(key)

	if err != nil {

		return "", err

	}

	decodedMsg, _ := hex.DecodeString(text)

	iv := make([]byte, aes.BlockSize)

	msg := decodedMsg

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(msg, msg)

	unpadMsg, err := unpad(msg)

	if err != nil {

		return "", err

	}

	return string(unpadMsg), nil

}

func test() {

	//key := []byte("0123456789abcdea")
	key := []byte("linkingthingrcom")

	//encryptText, _ := Encrypt(key, "123456")
	encryptText, _ := Encrypt(key, "admin")
	//fmt.Println(encryptText)
	finalMsg, err := base64.StdEncoding.DecodeString(encryptText)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(finalMsg))

	rawText, err := Decrypt(key, encryptText)
	//rawText, err := Decrypt(key, "AAAAAAAAAAAAAAAAAAAAAC7vq2/bL/QY6jRM5xap0ZM=")
	if err != nil {
		panic(err)
	}

	fmt.Println("text ", rawText)

}

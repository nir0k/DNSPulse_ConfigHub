package grpcserver

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)


func Encrypt(plainText string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    plainData := PKCS7Pad([]byte(plainText), block.BlockSize())

    cipherText := make([]byte, aes.BlockSize+len(plainData))
    iv := cipherText[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCBCEncrypter(block, iv)
    stream.CryptBlocks(cipherText[aes.BlockSize:], plainData)

    return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(cipherText string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    cipherData, err := base64.StdEncoding.DecodeString(cipherText)
    if err != nil {
        return "", err
    }

    if len(cipherData) < aes.BlockSize {
        return "", errors.New("cipherText too short")
    }

    iv := cipherData[:aes.BlockSize]
    cipherData = cipherData[aes.BlockSize:]

    stream := cipher.NewCBCDecrypter(block, iv)
    stream.CryptBlocks(cipherData, cipherData)

    data, err := PKCS7Unpad(cipherData)
    if err != nil {
        return "", err
    }

    return string(data), nil
}

func PKCS7Pad(data []byte, blockLen int) []byte {
    padding := blockLen - len(data)%blockLen
    padText := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(data, padText...)
}

func PKCS7Unpad(data []byte) ([]byte, error) {
    length := len(data)
    if length == 0 {
        return nil, errors.New("input[]byte is empty")
    }
    pad := int(data[length-1])
    if pad < 1 || pad > 16 {
        return nil, errors.New("padding size is wrong")
    }
    for i := 0; i < pad; i++ {
        if data[length-i-1] != byte(pad) {
            return nil, errors.New("invalid padding")
        }
    }
    return data[:length-pad], nil
}
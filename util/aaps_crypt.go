package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

// the following values are from AAPS to ensure everything is identical:
// https://github.com/nightscout/AndroidAPS/blob/219bdba21531c8f9d5df0ebaf5a7a3821c179d9a/core/src/main/java/info/nightscout/androidaps/utils/CryptoUtil.kt#L26
const (
	// IvLengthByte IV is equivalent to `nonce` here
	IvLengthByte     = 12
	TagLengthBit     = 128
	AESKeySizeBit    = 256
	PBKDF2Iterations = 50000
	SaltSizeByte     = 32
)

func Sha256(data []byte) string {
	sha := sha256.New()
	sha.Write(data)
	return hex.EncodeToString(sha.Sum(nil))
}

func Hmac256(message []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSizeByte)
	_, err := rand.Read(salt)
	return salt, err
}

func DeriveKey(passphrase []byte, salt []byte) []byte {
	return pbkdf2.Key(passphrase, salt, PBKDF2Iterations, AESKeySizeBit/8, sha1.New)
}

//ParseAAPSEncoding parses the `content` key in settings export for the nonce and cipher text data.
func ParseAAPSEncoding(content string) (nonce []byte, cipherText []byte, err error) {
	decodedData, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, nil, err
	}

	// AAPS includes the nonce in a custom header.
	// breakdown:
	// byte[0]                  = nonceLength
	// next `nonceLength` bytes = nonce
	// rest of bytes            = cipher text
	buffer := bytes.NewReader(decodedData)
	nonceLength, err := buffer.ReadByte()
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, nonceLength)
	_, err = buffer.Read(nonce)
	if err != nil {
		return nil, nil, err
	}

	cipherText = make([]byte, buffer.Len())
	_, err = buffer.Read(cipherText)
	if err != nil {
		return nil, nil, err
	}

	return nonce, cipherText, nil
}

func Encrypt(passphrase []byte, salt []byte, rawData []byte) ([]byte, error) {
	key := DeriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, IvLengthByte)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCMWithTagSize(block, TagLengthBit/8)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, rawData, nil)

	// create the final output, with the AAPS nonce header
	buf := make([]byte, 1, 1+len(nonce)+len(ciphertext))
	buf[0] = byte(len(nonce))
	buf = append(buf, nonce...)
	buf = append(buf, ciphertext...)
	return []byte(base64.StdEncoding.EncodeToString(buf)), nil
}

func Decrypt(passphrase []byte, salt []byte, encodedData string) ([]byte, error) {
	key := DeriveKey(passphrase, salt)

	nonce, cipherText, err := ParseAAPSEncoding(encodedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

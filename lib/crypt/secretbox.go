package crypt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
)

type OmniCrypter struct {
	SecretKey [32]byte
}

// NewOmniCrypter takes a 32byte hex encoded string
func NewOmniCrypter(key string) (*OmniCrypter, error) {
	var secretKey [32]byte
	keyAsBytes, decodeErr := hex.DecodeString(key)

	if decodeErr != nil {
		return nil, decodeErr
	}

	if len(keyAsBytes) < 32 {
		return nil, errors.New("key length insufficient")
	}

	copy(secretKey[:], keyAsBytes)

	return &OmniCrypter{secretKey}, nil
}

func (self *OmniCrypter) Encrypt(msg []byte) ([]byte, error) {
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		var nilSlice []byte
		return nilSlice, err
	}

	// This encrypts "hello world" and appends the result to the nonce.
	return secretbox.Seal(nonce[:], msg, &nonce, &self.SecretKey), nil
}

func (self *OmniCrypter) Decrypt(encrypted []byte) ([]byte, error) {
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &self.SecretKey)

	if !ok {
		return decrypted, errors.New("decryption error")
	} else {
		return decrypted, nil
	}
}

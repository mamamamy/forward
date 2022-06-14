package util

import (
	"crypto/sha256"
)

type streamCipherUtil struct{}

var StreamCipher streamCipherUtil

type StreamEncrypter struct {
	key                []byte
	lastPlainTextBlock []byte
	keyBlock           []byte
	keyBlockI          int
}

func (streamCipherUtil) NewStreamEncrypter(key []byte) *StreamEncrypter {
	initKeyBlock := sha256.Sum256(key)
	initLastPlainTxtBlock := sha256.Sum256(initKeyBlock[:])
	return &StreamEncrypter{
		key:                key,
		lastPlainTextBlock: initLastPlainTxtBlock[:],
		keyBlock:           initKeyBlock[:],
	}
}

func (s *StreamEncrypter) XORKeyStream(dst, src []byte) {
	for i := range src {
		if s.keyBlockI >= sha256.Size {
			h := sha256.New()
			h.Write(s.key)
			h.Write(s.keyBlock)
			h.Write(s.lastPlainTextBlock)
			s.keyBlock = h.Sum(nil)
			s.keyBlockI = 0
		}
		s.lastPlainTextBlock[s.keyBlockI] += src[i]
		dst[i] = src[i] ^ s.keyBlock[s.keyBlockI]
		s.keyBlockI++
	}
}

type StreamDecrypter struct {
	key                []byte
	lastPlainTextBlock []byte
	keyBlock           []byte
	keyBlockI          int
}

func (streamCipherUtil) NewStreamDecrypter(key []byte) *StreamDecrypter {
	initKeyBlock := sha256.Sum256(key)
	initLastPlainTxtBlock := sha256.Sum256(initKeyBlock[:])
	return &StreamDecrypter{
		key:                key,
		lastPlainTextBlock: initLastPlainTxtBlock[:],
		keyBlock:           initKeyBlock[:],
	}
}

func (s *StreamDecrypter) XORKeyStream(dst, src []byte) {
	for i := range src {
		if s.keyBlockI >= sha256.Size {
			h := sha256.New()
			h.Write(s.key)
			h.Write(s.keyBlock)
			h.Write(s.lastPlainTextBlock)
			s.keyBlock = h.Sum(nil)
			s.keyBlockI = 0
		}
		dst[i] = src[i] ^ s.keyBlock[s.keyBlockI]
		s.lastPlainTextBlock[s.keyBlockI] += dst[i]
		s.keyBlockI++
	}
}

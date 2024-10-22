package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

func MD5(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

func SHA(s string) string {
	sha := sha1.New()
	sha.Write([]byte(s))
	return hex.EncodeToString(sha.Sum(nil))
}

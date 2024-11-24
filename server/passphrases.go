package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

func WriteHashFromPassphrase(passphrase string) error {
	return writeHash(hashPassphrase(passphrase, generateSalt()))
}

func writeHash(hash string) error {
	hashfile_dir := filepath.Dir(HASH_FILE)
	err := os.MkdirAll(hashfile_dir, 0700)

	if err != nil {
		return err
	}
	return os.WriteFile(HASH_FILE, []byte(hash), 0600)
}

func CheckHash(passphrase string) (bool, error) {
	storedHashAndSalt, err := os.ReadFile(HASH_FILE)
	storedSalt := storedHashAndSalt[64:]
	hash := hashPassphrase(passphrase, string(storedSalt))
	return hash == string(storedHashAndSalt), err
}

func hashPassphrase(passphrase, salt string) string {
	prevHash := passphrase
	iterations := PowInts(2, 16)

	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write([]byte(prevHash + salt))
		prevHash = hex.EncodeToString(h.Sum(nil))
	}
	return prevHash + salt
}

func PowInts(x, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	y := PowInts(x, n/2)
	if n%2 == 0 {
		return y * y
	}
	return x * y * y
}

func generateSalt() string {
	r := rand.New(rand.NewSource(genTrulyRandom()))
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%f", r.Float64())))
	return hex.EncodeToString(h.Sum(nil))
}

func genTrulyRandom() int64 {
	file, _ := os.Open("/dev/urandom")
	defer file.Close()

	b := make([]byte, 256)
	file.Read(b)

	return int64(binary.BigEndian.Uint64(b))
}
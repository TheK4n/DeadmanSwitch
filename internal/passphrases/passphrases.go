package passphrases

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"math"
	"os"
)

func CheckHash(passphrase string, hash string) bool {
	salt := extractSalt(hash)
	calculatedHash := calculateHash(passphrase, salt)
	return hash == calculatedHash
}

func extractSalt(hash string) string {
	return hash[64:]
}

func HashSaltPassphrase(passphrase string) string {
	return calculateHash(passphrase, generateSalt())
}

func calculateHash(passphrase string, salt string) string {
	prevHash := passphrase
	iterations := int(math.Pow(2, 16))

	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write([]byte(prevHash + salt))
		prevHash = hex.EncodeToString(h.Sum(nil))
	}
	return prevHash + salt
}

func generateSalt() string {
	r := rand.New(rand.NewSource(generateStrongRandom()))
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%f", r.Float64())))
	return hex.EncodeToString(h.Sum(nil))
}

func generateStrongRandom() int64 {
	file, _ := os.Open("/dev/urandom")
	defer file.Close()

	b := make([]byte, 256)
	file.Read(b)

	return int64(binary.BigEndian.Uint64(b))
}
package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,32}$`)

func newAuthID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || utf8.RuneCountInString(value) > 191 {
		return false
	}
	parts := strings.Split(value, "@")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[0]) <= 64 &&
		strings.Contains(parts[1], ".") && !strings.HasPrefix(parts[1], ".") && !strings.HasSuffix(parts[1], ".")
}

func validPassword(value string, minimum int) bool {
	length := utf8.RuneCountInString(value)
	return length >= minimum && length <= 128
}

func validAuthToken(value string) bool {
	length := utf8.RuneCountInString(value)
	return length >= 20 && length <= 512
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	return fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

// Preserve the PHC encoding written by the original Node argon2 library.
func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	parameters := strings.Split(parts[3], ",")
	if len(parameters) != 3 {
		return false
	}
	values := make([]uint64, 3)
	for index, prefix := range []string{"m=", "t=", "p="} {
		if !strings.HasPrefix(parameters[index], prefix) {
			return false
		}
		value, err := strconv.ParseUint(strings.TrimPrefix(parameters[index], prefix), 10, 32)
		if err != nil {
			return false
		}
		values[index] = value
	}
	memory, iterations, parallelism := values[0], values[1], values[2]
	if parallelism < 1 || parallelism > 16 || memory < 8*parallelism || memory > 262144 || iterations < 1 || iterations > 10 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) < 16 || len(want) > 64 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func randomAuthToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashAuthToken(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func limitedAuthString(value string, length int) *string {
	if value == "" {
		return nil
	}
	runes := []rune(value)
	if len(runes) > length {
		value = string(runes[:length])
	}
	return &value
}

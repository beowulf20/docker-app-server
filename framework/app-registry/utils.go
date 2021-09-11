package app_registry

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

var hasher = sha256.New()

func calcHash(s string) (string, error) {
	defer hasher.Reset()
	reader := strings.NewReader(s)
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

var ErrStringContainsSpecialChars = errors.New("string contains invalid characters")
var appNameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func validateStringSpecialCharacters(s string) error {
	return validation.Validate(s,
		validation.Required,
		validation.Length(2, 50),
		validation.NewStringRuleWithError(appNameRegex.MatchString, is.ErrAlphanumeric),
	)
}

func CalculateHash(obj interface{}) (string, error) {
	return calcHash(fmt.Sprintf("%v", obj))
}

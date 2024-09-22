package util

import (
	"crypto/rand"
	"log"
	"math/big"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func FormatName(title string) string {
	name := regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(title, "")
	name = strings.ReplaceAll(name, " ", "_")
	return strings.ToLower(name)
}

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*")
	str := make([]rune, length)
	for i := range str {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatal("Could not generate random string")
		}
		str[i] = letters[value.Int64()]
	}
	return string(str)
}

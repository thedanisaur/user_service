package util

import (
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

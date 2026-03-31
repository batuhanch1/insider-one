package utils

import (
	"fmt"
	"runtime"
)

func GetStackTrace(err error) string {
	stack := make([]byte, 4<<10)
	length := runtime.Stack(stack, false)
	return fmt.Sprintf("%v %s\n", err, stack[:length])
}

func InterfaceToString(data interface{}) (value string, isSuccess bool) {
	if data == nil {
		return "", false
	}

	if value, isSuccess = data.(string); !isSuccess {
		return "", false
	}

	return value, true
}

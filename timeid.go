package utils

import (
	"strconv"
	"time"
)

func TimeID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

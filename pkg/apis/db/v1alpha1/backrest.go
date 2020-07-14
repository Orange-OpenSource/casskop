package v1alpha1

import (
	"fmt"
	"strconv"
)

func ProgressPercentage(progress float64) string {
	return fmt.Sprintf("%v%%", strconv.Itoa(int(progress*100)))
}

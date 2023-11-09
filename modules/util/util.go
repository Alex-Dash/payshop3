package util

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strconv"
)

func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func ToFixedDecimal(num int, decimal int) string {
	p := math.Pow10(decimal)
	n := float64(num) / p
	return strconv.FormatFloat(n, 'g', decimal, 64)
}

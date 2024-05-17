package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const dayFormat = "2006-01-02"

const dayTimeFormat = "2006-01-02 15:04:05"

func INFO(msg ...any) {
	if len(msg) > 0 {
		msg = append([]any{fmt.Sprintf("[INFO] %s ", time.Now().Format(dayTimeFormat))}, msg...)
	}
	m := fmt.Sprint(msg...)
	fmt.Println(m)
	writeFile(m)
}

func DEBUG(msg ...any) {
	if len(msg) > 0 {
		msg = append([]any{fmt.Sprintf("[DEBUG] %s ", time.Now().Format(dayTimeFormat))}, msg...)
	}
	if os.Getenv("GIN_MODE") != gin.ReleaseMode {
		fmt.Println(msg...)
	}
}

func ERROR(msg ...any) {
	if len(msg) > 0 {
		msg = append([]any{fmt.Sprintf("[ERROR] %s ", time.Now().Format(dayTimeFormat))}, msg...)
	}
	m := fmt.Sprint(msg...)
	fmt.Println(m)
	writeFile(m)
}

func writeFile(msg string) {
	if _, err := os.Stat(getFile()); os.IsNotExist(err) {
		os.WriteFile(getFile(), []byte(""), os.ModePerm)
	}
	os.WriteFile(getFile(), []byte(msg), os.ModeAppend)
}

func getFile() string {
	return "./logs/" + time.Now().Format(dayFormat) + ".log"
}

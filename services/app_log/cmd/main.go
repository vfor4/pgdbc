package main

import (
	"fmt"
	stlog "log"
	applog "main/services/app_log"
)

func main() {
	applog.Run("logs")

	host, port := "localhost", "4000"
	serviceAddress := fmt.Sprintf("http://%v:%v", host, port)
}

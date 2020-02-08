package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	file := "./access.log"

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("open file error: %s", err.Error()))
	}

	for {
		f.WriteString("127.0.0.1 - - [21/Dec/2015:20:22:14 +0800] http \"GET /phpinfo.php HTTP/1.1\" 200 12704 \"-\" \"KeepAliveClient\" \"-\" 1.005 1.854\n")
		time.Sleep(time.Millisecond * 100)
	}
}
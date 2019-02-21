package main

import (
	"fmt"
	"os"
	"time"

	"github.com/windmilleng/tilt/internal/cli"
)

func main() {
	err := new(cli.SyncletCmd).Register().Execute()
	if err != nil {
		fmt.Println(err)
		time.Sleep(3600 * time.Second)
		os.Exit(1)
	}
}

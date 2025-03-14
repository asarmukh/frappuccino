package helper

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"frappuccino/utils"
)

var counter uint64

// инкрементит безопасно id
func GenerateID() uint64 {
	return atomic.AddUint64(&counter, 1)
}

func PrintUsage() {
	fmt.Println(`$ ./hot-coffee --help
Coffee Shop Management System

Usage:
  hot-coffee [--port <N>] [--dir <S>] 
  hot-coffee --help

Options:
  --help       Show this screen.
  --port N     Port number.
  --dir S      Path to the data directory.`)
}

func CreateNewDir(dir string) {
	if !utils.IsValidDir(dir) {
		log.Fatal("Error: incorrect directory")
		os.Exit(1)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	fileArray := [3]string{"orders.json", "inventory.json", "menu_items.json"}

	for i := 0; i < len(fileArray); i++ {
		filePath := dir + "/" + fileArray[i]

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			file, err := os.Create(filePath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
		}
	}
}

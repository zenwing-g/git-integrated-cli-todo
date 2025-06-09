package main

import (
	"fmt"
	"os"

	"todo/handlers"
	"todo/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	switch os.Args[1] {
	case "init":
		handlers.HandleInit(os.Args[2:])
	case "rm":
		handlers.HandleDelete()
	case "add":
		if err := storage.InitTaskID(); err != nil {
			fmt.Println("[-] Failed to initialize taskID:", err)
			return
		}
		handlers.HandleAdd()
	case "ls":
		handlers.HandleList(os.Args[2:])
	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Try: todo [init|ls|add|done|rm]")
	}
}

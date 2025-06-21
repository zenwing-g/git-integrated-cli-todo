package main

import (
	"fmt"
	"os"

	"todo/handlers"
)

func main() {
	fmt.Println("[*] Starting todo CLI...")

	// basic CLI routing
	if len(os.Args) < 2 {
		fmt.Println("[*] Usage:\ntodo [init|ls|add|done|rm]")
		return
	}

	switch os.Args[1] {
	case "init":
		handlers.HandleInit(os.Args[2:])
	case "rm":
		handlers.HandleDelete()
	case "add":
		handlers.HandleAdd()
	case "ls":
		handlers.HandleList(os.Args[2:])
	case "done":
		handlers.HandleCompleted(os.Args[2:])
	default:
		fmt.Println("[x] Unknown command:", os.Args[1])
		fmt.Println("Try: todo [init|ls|add|done|rm]")
	}
}

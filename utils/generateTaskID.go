package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"todo/constants"
)

// generate unique ID, avoid collisions
func GenerateTaskID(length int) string {
	filePath := constants.IdsFilePath

	for {
		id := make([]byte, length)
		max := big.NewInt(int64(len(constants.CharSet)))

		for i := range id {
			n, err := rand.Int(rand.Reader, max)
			if err != nil {
				panic(err)
			}
			id[i] = constants.CharSet[n.Int64()]
		}
		ID := string(id)

		if err := CheckIfIDsFileExists(filePath); err != nil {
			fmt.Println("[x] Error ensuring ids.txt exists:", err)
			return ""
		}

		if !CheckIfIDExists(filePath, ID) {
			AppendIDToFile(filePath, ID)
			return ID
		}
	}
}

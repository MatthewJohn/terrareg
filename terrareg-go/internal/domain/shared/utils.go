package shared

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateID generates a random ID string
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// GenerateRandomID generates a random ID string (alias for GenerateID)
func GenerateRandomID() string {
	return GenerateID()
}

// GenerateIntID generates a random integer ID
func GenerateIntID() int {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	id := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	if id < 0 {
		id = -id
	}
	return id
}

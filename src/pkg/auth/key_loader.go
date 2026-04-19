package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadPublicKeys loads public keys from env var JSON or from a specific directory.
// Returns a map of kid -> publicKeyPEM.
func LoadPublicKeys(envJSON string, keysDir string) (map[string]string, error) {
	keys := make(map[string]string)

	// 1. Try to load from JSON string in environment variable
	if envJSON != "" {
		err := json.Unmarshal([]byte(envJSON), &keys)
		if err == nil && len(keys) > 0 {
			log.Println("Loaded public keys from environment JSON.")
			return keys, err
		}
		log.Printf("Failed to parse public keys from env JSON or it was empty. Error: %v\n", err)
	}

	// 2. Fallback to reading all .pem files in the keys directory
	log.Printf("Scanning keys directory %s for public keys...\n", keysDir)
	entries, err := os.ReadDir(keysDir)
	if err != nil {
		if errorsIsNotExist(err) {
			return nil, fmt.Errorf("keys directory %s does not exist", keysDir)
		}
		return nil, fmt.Errorf("failed to read keys directory: %v", err)
	}

	foundKeys := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pem" {
			continue
		}

		// Skip private keys to avoid accidentally loading them into verifiers
		if strings.Contains(entry.Name(), "private") {
			continue
		}

		path := filepath.Join(keysDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Could not read key file %s: %v\n", path, err)
			continue
		}

		// Use filename without extension as the default 'kid' for file-based keys
		kid := strings.TrimSuffix(entry.Name(), ".pem")
		
		// If someone overrides kid internally inside file they could mapping it, 
		// but using filename is standard for basic rotations
		keys[kid] = string(data)
		foundKeys++
	}

	if foundKeys == 0 {
		return nil, fmt.Errorf("no public keys found in %s and env JSON was empty", keysDir)
	}

	log.Printf("Loaded %d public keys from directory.\n", foundKeys)
	return keys, nil
}

func errorsIsNotExist(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist)
}

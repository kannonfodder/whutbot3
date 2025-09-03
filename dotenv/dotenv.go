package dotenv

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// Load reads key=value lines from the given .env file and sets environment
// variables for any keys that are not already set in the environment.
func Load(path string) {
	f, err := os.Open(path)
	if err != nil {
		// no .env present — nothing to do
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// strip surrounding quotes if any
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		// only set env var if it's not already set in the environment
		if os.Getenv(key) == "" {
			err := os.Setenv(key, val)
			if err != nil {
				log.Printf("failed to set env %s: %v", key, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading %s: %v", path, err)
	}
}

package messages

import (
	"strings"
)

func parseCommand(input string) (command string, arguments string) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) > 1 {
		return strings.ToLower(parts[0]), strings.ToLower(parts[1])
	}
	return strings.ToLower(parts[0]), ""
}

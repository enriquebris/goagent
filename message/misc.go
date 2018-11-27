package message

import (
	"fmt"
)

// getKeyValueTable returns a string formatted as a table
func GetKeyValueTable(param ...string) string {
	// get key max length
	maxLen := 0
	for i := 0; i < len(param); i = i + 2 {
		if maxLen < len(param[i]) {
			maxLen = len(param[i])
		}
	}

	ret := ""
	for i := 0; i < len(param); i = i + 2 {
		ret += fmt.Sprintf(
			"\n%v%v\t\t%v\n",
			param[i],                        // command
			GetSpaces(maxLen-len(param[i])), // extra blank spaces to match all columns
			param[i+1],                      // command description
		)
	}

	return ret
}

// getSpaces returns a string having only "total" blank spaces
func GetSpaces(total int) string {
	ret := ""
	for i := 0; i < total; i++ {
		ret += " "
	}

	return ret
}

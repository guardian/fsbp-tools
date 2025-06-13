package common

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func UserConfirmation() bool {

	buf := bufio.NewReader(os.Stdin)
	fmt.Println("Press 'y', to confirm, and enter to continue. Otherwise, hit enter to exit.")
	fmt.Print("> ")
	input, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println("Error reading input: " + err.Error())
	}
	return strings.ToLower(strings.TrimSpace(string(input))) == "y"
}

func ExitOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

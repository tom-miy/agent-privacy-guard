package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func readInput(path string) (string, error) {
	var b []byte
	var err error
	if path == "" || path == "-" {
		b, err = io.ReadAll(os.Stdin)
	} else {
		b, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

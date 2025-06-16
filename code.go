package main

import (
	"fmt"
    "os/exec"
)

func manipulateImage(filename string) error {
	// run the command `convert filename output.png`:
	cmd := exec.Command("convert", filename, "output.png")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("command failed: %v, output: %s", err, string(output))
    }
    fmt.Printf("Command output: %s\n", string(output))
    return nil
}


func main() {
	err := manipulateImage("; touch input.jpg")
	if err != nil {
		panic(err)
	}
}
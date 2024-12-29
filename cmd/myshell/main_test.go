package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pwd)
	cmd := exec.Command("cat", "'abc renu.txt'")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(string(output))
	}
}

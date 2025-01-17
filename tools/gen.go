//go:build ignore

package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "install",
		"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to install protoc-gen-go: %v\n", err)
		return
	}

	cmd = exec.Command("go", "generate", "./...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to generate proto files: %v\n", err)
	}
}

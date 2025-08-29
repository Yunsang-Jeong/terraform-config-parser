package main

import (
	"context"
	"os"

	"github.com/Yunsang-Jeong/terraform-config-parser/cmd"
)

func main() {
	ctx := context.Background()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(-1)
	}
}

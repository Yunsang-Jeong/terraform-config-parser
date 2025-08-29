package main

import (
	"context"
	"os"
	"terraform-config-parser/cmd"
)

func main() {
	ctx := context.Background()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(-1)
	}
}

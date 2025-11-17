/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"osom/cmd"
	config "osom/pkg"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	config.Init()

	cmd.Execute()
}

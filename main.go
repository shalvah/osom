/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"osom/cmd"
	pkg "osom/pkg"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	pkg.InitConfig()
	pkg.InitLogging()

	cmd.Execute()
}

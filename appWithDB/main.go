package main

import (
	"fmt"

	"./dcmd"

	"static"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	fmt.Println(static.Banner("Welcome!"))

	dcmd.Execute()
}

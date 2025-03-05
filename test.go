package main

import (
	"fmt"
	log "log"
	"os"

	"github.com/joho/godotenv"
)

func importEnv(fileName, varName string) (variable string) {
	err := godotenv.Load(fileName)
	if err != nil {
		log.Fatalf("Ошибка импорта файла %v", err)
	}

	variable = os.Getenv(varName)
	if variable == "" {
		log.Fatalf("Переменная %v не найдена.", variable)
	}
	fmt.Println("Переменная ", varName, " из файла ", fileName, " импортирована!")
	return
}

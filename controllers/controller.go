package controllers

import (
	"log"
	"os/exec"
)

func GenerateId() []byte {
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	return newUUID
}

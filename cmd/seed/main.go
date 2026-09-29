package main

import (
	"fmt"
	"log"

	bcryptadapter "barbershop/internal/adapters/bcrypt"
)

func main() {
	h := bcryptadapter.NewHasher()
	hash, err := h.Hash("admin123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
}

package main

import (
	"fmt"

	"github.com/henryfool91/pet-sso/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
	//TODO: инит логгера

	//TODO: инит апп

	//TODO: запуск grpc сервера

}

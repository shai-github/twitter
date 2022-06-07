package main

import (
	"encoding/json"
	"os"
	"proj1/server"
	"strconv"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	argLength := len(os.Args)
	if argLength == 2 {
		consumerCount, _ := strconv.Atoi(os.Args[1])
		cfg := server.Config{Encoder: encoder, Decoder: decoder, Mode: "p", ConsumersCount: consumerCount}
		server.Run(cfg)
	} else {
		cfg := server.Config{Encoder: encoder, Decoder: decoder, Mode: "s"}
		server.Run(cfg)
	}
}

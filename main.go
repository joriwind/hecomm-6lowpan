package main

import (
	"context"

	"github.com/joriwind/hecomm-6lowpan/cisixlowpan"
)

func main() {
	channel := make(chan cisixlowpan.Message)
	cisixlowpan.NewServer(context.Background(), channel)
}

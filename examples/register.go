package main

import (
	"fmt"
	"sync"
	pulsatio "github.com/Weinsen/pulsat.io-go"
)

func Print(data string) {
	fmt.Printf("CALLBACK: %s\n", data)
}

func main() {

	p := pulsatio.New("1", "http://localhost:3000")
	p.SetCallback("connection", Print)
	p.SetCallback("heartbeat", Print)
	p.Start()


	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()

}
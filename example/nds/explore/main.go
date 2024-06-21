package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sukus21/nintil/nds"
	"github.com/sukus21/nintil/util"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify ROM as command line parameter")
	}

	f := util.Must1(os.Open(os.Args[1]))
	defer f.Close()
	rom := util.Must1(nds.OpenROM(f))

	for {
		fmt.Print("Enter ROM address: ")
		addr := uint32(0)
		_, err := fmt.Scanf("%X\n", &addr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		entry := rom.WhatsHere(addr)
		if entry == nil {
			fmt.Println("address not mapped")
		} else {
			fmt.Printf("%s [%X]\n", entry.Name(), addr-entry.From())
		}
	}
}

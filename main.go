package main

import (
	modhandler "gitlab.com/enomics/modbus-client/enom-modbus"
)

func main() {
	ch := make(chan modhandler.Request, 100)

	go modhandler.Run(ch, modhandler.NewConfig())

	for i := 0; i < 10; i++ {
		// s := strconv.Itoa(i)
		ch <- modhandler.Request{
			Cb: func(res []byte, err error) {
				if err != nil {
					println(err)
					return
				}

				println(res)

			}}
	}

	select {}
}

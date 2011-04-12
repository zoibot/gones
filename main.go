package main

import (
    "os"
    "./gones"
)

func main() {
    m := gones.MakeMachine(os.Args[1])
    m.Run()
    //fmt.Printf("%v\n", m)
}

package main

import (
    "fmt"
    "os"
    "./gones"
)

func main() {
    fmt.Printf("Hello, world!\n")
    m := gones.MakeMachine(os.Args[1])
    m.Run()
    //fmt.Printf("%v\n", m)
}

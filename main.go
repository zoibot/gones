package main

import (
    "fmt"
    "./gones"
)

func main() {
    fmt.Printf("Hello, world!\n")
    m := gones.MakeMachine("test")
    fmt.Printf("%v\n", m)
}

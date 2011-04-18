package main

import (
    "os"
    "fmt"
    "bytes"
    "big"
    "strconv"
    "./gones"
)


func test(tfile string) {
    f, e := os.Open(tfile)
    if f == nil {
        fmt.Printf("Error opening test file: %v\n", e)
    }
    finfo, _ := f.Stat()
    sz := finfo.Size
    buf := make([]byte, sz)
    f.Read(buf)
    lines := bytes.Split(buf, []byte{'\n'}, -1)
    romname := lines[0]
    frames := make(chan []int)
    currentInput := make([]byte, 512)
    m := gones.MakeMachine(string(romname), frames,
        func() chan []byte {
            c := make(chan []byte)
            go func() {
                for {
                    c <- currentInput
                }
            }()
            return c
        }())
    go m.Run()
    for line := range lines[1:] {
        fs := bytes.Fields(lines[line])
        switch string(fs[0]) {
        case "wait":
            length, _ := strconv.Atoi(string(fs[1]))
            //fmt.Printf("waiting %v\n", length)
            for i := 0; i < length; i++ {
                <-frames
            }
        case "screen":
            //fmt.Printf("screenshot\n")
            frame := <-frames
            gones.SaveImage("test/test.png", frame)
        case "test_image":
            frame := <-frames
            hash := big.NewInt(0)
            hash.SetBytes(gones.HashImage(frame))
            if bytes.Compare(fs[1], []byte(fmt.Sprintf("%x", hash))) != 0 {
                fmt.Printf("Fail %x\n", hash)
                gones.SaveImage("test/test.png", frame)
            } else {
                fmt.Printf("Pass!\n")
            }
        case "press":
            key, _ := strconv.Atoi(string(fs[1]))
            //fmt.Printf("pressing key %v\n", key)
            currentInput[key] = 1
            <-frames
            <-frames
            currentInput[key] = 0
            //case "test":
        }
    }
}

package main

import (
    "⚛sdl"
    "os"
    "fmt"
    "./gones"
    "path/filepath"
    "flag"
)

func sdlInput() chan []byte {
    c := make(chan []byte)
    go func () {
    for {
        c <- []byte(sdl.GetKeyState())
    }
    }()
    return c
}

func readStdin() chan byte {
    c := make(chan byte)
    go func () {
    b := make([]byte, 1)
    for {
        n, _ := os.Stdin.Read(b)
        if n != 0 {
            c <- b[0]
        }
    }
    }()
    return c
}

func main() {
    //set up command line options
    var inputFile, recordKeys, testFile string
    flag.StringVar(&inputFile, "input", "", "Specify an input file to use instead of keypresses")
    flag.StringVar(&testFile, "test", "", "Specify a test file to run")
    flag.StringVar(&recordKeys, "record", "", "Record keypresses for later playback")
    var suppressVideo bool
    flag.BoolVar(&suppressVideo, "novideo", false, "Disable video output for running in testing mode")
    flag.Parse()

    if testFile != "" {
        test(testFile)
    } else {
        run()
    }
}

func run() {
    //initialize video if we need to 
    var screen *sdl.Surface
    sdl.Init(sdl.INIT_VIDEO)
    screen = sdl.SetVideoMode(256, 240, 32, 0)
    sdl.WM_SetCaption("gones","")
    //set up channels for communicating with machine
    frames := make(chan []int)

    romfile := flag.Arg(0)
    romname := filepath.Base(flag.Arg(0))
    num := 0
    m := gones.MakeMachine(romfile, frames, sdlInput())

    video := false
    //run machine
    go m.Run()
    //start reading std input
    input := readStdin()
    for {
        select {
        case frame := <-frames:
            //new frame
            copy((*[256*240]int)(screen.Pixels)[:], frame)
            if video {
                gones.SaveImage(fmt.Sprintf("video/%s_%05d.png", romname, num), frame)
                num++
            }
            screen.Flip()
        case c := <-input:
            //char from stdin
            switch c {
                case 's':
                    gones.SaveImage("ss.png", (*[256*240]int)(screen.Pixels)[:])
            }
        case event := <-sdl.Events:
            switch e := event.(type) {
            case sdl.QuitEvent:
                sdl.Quit()
                os.Exit(0)
            case sdl.KeyboardEvent:
                kevent := event.(sdl.KeyboardEvent)
                if kevent.Type == sdl.KEYDOWN {
                    break
                }
                switch kevent.Keysym.Sym {
                case sdl.K_v:
                    video = !video
                    fmt.Printf("recording video: %v\n", video)
                    num = 0
                default:
                    m.Debug(kevent.Keysym.Sym)
                }
            }
        }
    }
}

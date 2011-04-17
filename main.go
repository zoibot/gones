package main

import (
    "⚛sdl"
    "os"
    "./gones"
)

func main() {
    sdl.Init(sdl.INIT_VIDEO)
    screen := sdl.SetVideoMode(256, 240, 32, 0)
    sdl.WM_SetCaption("gones","")
    frames := make(chan []int)
    //input := make(chan []byte)
    m := gones.MakeMachine(os.Args[1], frames)
    go m.Run()
    for {
        select {
        case frame := <-frames:
            //draw frame somehow
            copy((*[256*240]int)(screen.Pixels)[:], frame)
            screen.Flip()
        //case <-input:
            //input <- get key state or whatever
        case event := <-sdl.Events:
            switch e := event.(type) {
                case sdl.QuitEvent:
                    sdl.Quit()
                    os.Exit(0)
                case sdl.KeyboardEvent:
                    kevent := event.(sdl.KeyboardEvent)
                    m.Debug(kevent.Keysym.Sym)
            }
        }
    }
}

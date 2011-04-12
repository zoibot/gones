package gones

import "⚛sdl"
import "fmt"

type PPU struct {
    mach *Machine
    //SDL stuff
    screen *sdl.Surface
    cycle_count uint64
    //memory
    mem [0x4000]byte
    memBuf byte
    mirrorTable [0x4000]word
    latch bool
    pmask byte
    pstat byte
    pctrl byte
    //position
    xoff, fineX byte
    curSprs []*Sprite
    currentMirroring int
    vaddr, taddr word
    sl int
    cyc int
    objMem [0x100]byte
    objAddr word
}

type Sprite struct {
}

func makePPU(m *Machine) *PPU {
    sdl.Init(sdl.INIT_VIDEO);
    screen := sdl.SetVideoMode(256, 240, 32, 0)
    p := PPU{mach: m, screen: screen}
    for i := word(0); i < 0x4000; i++ {
        p.mirrorTable[i] = i
    }
    return &p
}

func (p *PPU) readRegister(num int) byte {
    return 0
}

func (p *PPU) writeRegister(num int, val byte)  {
}

func (p *PPU) run(cycles int) {
    select {
    case event := <-sdl.Events:
        fmt.Printf("%#v\n", event)
        switch e := event.(type) {
            case sdl.QuitEvent:
                sdl.Quit()
        }
    }
}


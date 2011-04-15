package gones

import (
"⚛sdl"
"fmt"
"os"
)

const (
    HORIZONTAL = iota
    VERTICAL
    FOUR_SCREEN
    SINGLE_LOWER
    SINGLE_UPPER
)

type PPU struct {
    mach *Machine
    cycles chan int
    //SDL stuff
    screen *sdl.Surface
    pixels []int
    cycleCount uint64
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
    curSprs [8]Sprite
    numSprs int
    currentMirroring int
    vaddr, taddr word
    sl int
    cyc int
    objMem [0x100]byte
    objAddr word
}

type Sprite struct {
    index int
    y, tile, attrs, x byte
}

func (s *Sprite) setSpr(index int, m []byte) {
    s.index = index
    s.y,s.tile,s.attrs,s.x = m[0], m[1], m[2], m[3]
}

func makePPU(m *Machine) *PPU {
    sdl.Init(sdl.INIT_VIDEO);
    screen := sdl.SetVideoMode(256, 240, 32, 0)
    p := PPU{mach: m, screen: screen}
    p.pixels = (*[256*240]int)(screen.Pixels)[:]
    for i := word(0); i < 0x4000; i++ {
        p.mirrorTable[i] = i
        p.mem[i] = 0xff
    }
    for i := int(0); i < 0x100; i++ {
        p.objMem[i] = 0xff
    }
    p.setMirroring(0x3000, 0x2000, 0xf00)
    p.currentMirroring = m.rom.mirror
    p.setNTMirroring(p.currentMirroring)
    p.cycles = make(chan int)
    return &p
}

func (p *PPU) dump() string {
    return fmt.Sprintf("CYC: %d SL: %d VADDR: %4X", p.cyc, p.sl, p.vaddr)
}

func (p *PPU) setMirroring(from word, to word, n word) {
    for i := word(0); i < n; i++ {
        p.mirrorTable[from+i] = to+i
    }
}

func (p *PPU) setNTMirroring(t int) {
    if t == p.currentMirroring { return }
    switch t {
	case VERTICAL:
		p.setMirroring(0x2000, 0x2000, 0x400);
		p.setMirroring(0x2400, 0x2400, 0x400);
		p.setMirroring(0x2800, 0x2000, 0x400);
        p.setMirroring(0x2c00, 0x2400, 0x400);
	case HORIZONTAL:
		p.setMirroring(0x2000, 0x2000, 0x400);
		p.setMirroring(0x2400, 0x2000, 0x400);
		p.setMirroring(0x2800, 0x2400, 0x400);
        p.setMirroring(0x2c00, 0x2400, 0x400);
	case SINGLE_LOWER:
		p.setMirroring(0x2000, 0x2000, 0x400);
		p.setMirroring(0x2400, 0x2000, 0x400);
		p.setMirroring(0x2800, 0x2000, 0x400);
        p.setMirroring(0x2c00, 0x2000, 0x400);
	case SINGLE_UPPER:
		p.setMirroring(0x2000, 0x2400, 0x400);
		p.setMirroring(0x2400, 0x2400, 0x400);
		p.setMirroring(0x2800, 0x2400, 0x400);
        p.setMirroring(0x2c00, 0x2400, 0x400);
        /*
	case SINGLE_THIRD:
		p.setMirroring(0x2000, 0x2800, 0x400);
		p.setMirroring(0x2400, 0x2800, 0x400);
		p.setMirroring(0x2800, 0x2800, 0x400);
        p.setMirroring(0x2c00, 0x2800, 0x400);
		break;
	case SINGLE_FOURTH:
		p.setMirroring(0x2000, 0x2c00, 0x400);
		p.setMirroring(0x2400, 0x2c00, 0x400);
		p.setMirroring(0x2800, 0x2c00, 0x400);
        p.setMirroring(0x2c00, 0x2c00, 0x400);
		break;
        */
	default:
		break;
    }
}

func (p *PPU) readRegister(num int) byte {
    ret := byte(0)
    switch num {
    case 2:
        ret = p.pstat
        p.pstat &= ^byte(1<<7)
        p.latch = false
        return ret
    case 4:
        return p.objMem[p.objAddr]
    case 7:
        if p.vaddr < 0x3f00 {
            ret = p.memBuf
            p.memBuf = p.getMem(p.vaddr)
        } else {
            p.memBuf = p.getMem(p.vaddr - 0x1000)
            ret = p.getMem(p.vaddr)
        }
        if p.pctrl & (1 << 2) != 0 {
            p.vaddr += 32
        } else {
            p.vaddr += 1
        }
        p.vaddr &= 0x3fff
        return ret
    }
    return 0
}

func (p *PPU) writeRegister(num int, val byte)  {
    switch num {
        case 0:
            p.pctrl = val
            p.taddr &= (^word(0x3 << 10))
            p.taddr |= word(val & 0x3) << 10
        case 1:
            p.pmask = val
        case 3:
            p.objAddr = word(val)
        case 4:
            p.objMem[p.objAddr] = val
            p.objAddr += 1
            p.objAddr &= 0xff
        case 5:
            if p.latch {
                p.taddr &= ^word(0x73e0)
                p.taddr |= word(val >> 3) << 5
                p.taddr |= word(val & 0x7) << 12
            } else {
                p.taddr &= ^word(0x1f)
                p.taddr |= word(val >> 3)
                p.xoff = val & 0x7
                p.fineX = val & 0x7
            }
            p.latch = !p.latch
        case 6:
            if p.latch {
                p.taddr &= ^word(0xff)
                p.taddr |= word(val)
                p.vaddr = p.taddr
            } else {
                p.taddr &= 0xff
                p.taddr |= word(val & 0x3f) << 8
            }
            p.latch = !p.latch
        case 7:
            p.setMem(p.vaddr, val)
            if p.pctrl & (1<<2) != 0 {
                p.vaddr += 32
            } else {
                p.vaddr += 1
            }
            p.vaddr &= 0x3fff
    }
}

func (p *PPU) getMem(addr word) byte {
    switch true {
        case addr < 0x2000:
            return p.mach.rom.chr_rom[(addr&0x1000)>>12][addr&0xfff]
        case addr < 0x3000:
            return p.mem[p.mirrorTable[addr]]
        case addr < 0x3f00:
            return p.getMem(addr - 0x1000)
        default:
            if addr & 0xf == 0 { addr = 0 }
            return p.mem[0x3f00 + (addr & 0x1f)]
    }
    return 0
}

func (p *PPU) setMem(addr word, val byte) {
    switch true {
        case addr < 0x2000:
            p.mach.rom.chr_rom[(addr&0x1000)>>12][addr&0xfff] = val
        case addr < 0x3f00:
            p.mem[p.mirrorTable[addr]] = val
        default:
            if addr & 0xf == 0 { addr = 0 }
            p.mem[0x3f00 + (addr & 0x1f)] = val
        }
}

func (p *PPU) newScanline() {
    fineY := (p.vaddr & 0x7000) >> 12
    if fineY == 7 {
        if p.vaddr & 0x3ff >= 0x3c0 {
            p.vaddr &= ^word(0x3ff)
        } else {
            p.vaddr += 0x20
            if p.vaddr & 0x3ff >= 0x3c0 {
                p.vaddr &= ^word(0x3ff)
                p.vaddr ^= 0x800
            }
        }
    }
    p.vaddr &= ^word(0x741f)
    p.vaddr |= p.taddr & 0x1f
    p.vaddr |= p.taddr & 0x400
    p.vaddr |= (fineY+1)&7 << 12
    p.fineX = p.xoff
    //sprites
    p.numSprs = 0
    curY := byte(p.sl-1)
    s := Sprite{}
    for i := 0; i < 64; i++ {
        (&s).setSpr(i,p.objMem[i*4:i*4+4])
        if s.y <= curY && (curY < s.y+8 || (p.pctrl & (1<<5)!=0 && curY< s.y+16)) {
            p.curSprs[p.numSprs] = s
            p.numSprs++
            if p.numSprs == 8 {
                break
            }
        }
    }
}

func (p *PPU) doVblank(renderingEnabled bool) {
    cycles := int(p.mach.cpu.cycleCount * 3 - p.cycleCount)
    p.pstat &= ^byte(1<<7)
    if 341-p.cyc > cycles {
        p.cyc += cycles
        p.cycleCount += uint64(cycles)
    } else {
        p.cycleCount += uint64(341 - p.cyc)
        p.cyc = 0
        p.sl += 1
        if renderingEnabled {
            p.vaddr = p.taddr
            p.fineX = p.xoff
        }
    }
}

func (p *PPU) renderPixels(x byte, y byte, num byte) {
    bgEnabled := p.pmask & (1<<3) != 0
    spriteEnabled := p.pmask & (1<<4) != 0
    fineY := (p.vaddr >> 12) & 7
    xoff := p.cyc
    basePtAddr := word(0x0)
    if p.pctrl & (1<<4) != 0 {
        basePtAddr = 0x1000
    }
    baseSprAddr := word(0x0)
    if p.pctrl & (1<<3) != 0 {
        baseSprAddr = 0x1000
    }
    for num != 0 {
        ntAddr := 0x2000 | (p.vaddr & 0xfff)
        atBase := (ntAddr & ^word(0x3ff)) + 0x3c0
        ntVal := word(p.getMem(ntAddr))
        ptAddr := (ntVal << 4) + basePtAddr
        row := (ntAddr >> 6) & 1
        col := (ntAddr & 2) >> 1
        atVal := p.getMem(atBase + ((ntAddr & 0x1f)>>2) + ((ntAddr & 0x3e0) >> 7) * 8)
        atVal >>= 4 * row + 2 * col
        atVal &= 3
        atVal <<=2
        hi := p.getMem(ptAddr + 8 + fineY)
        lo := p.getMem(ptAddr + fineY)
        hi >>= (7-p.fineX)
        hi &= 1
        hi <<= 1
        lo >>= (7-p.fineX)
        lo &= 1
        coli := word(0x3f00)
        if hi|lo != 0 && bgEnabled && !(xoff < 8 && (p.pmask & 2 == 0)) {
            coli |= word(atVal | hi | lo)
        }
        //TODO sprites
        if spriteEnabled && !(xoff < 8 && (p.pmask & 4 == 0)) {
            cur := Sprite{}
            for i := 0; i < p.numSprs; i++ {
                if p.curSprs[i].x <= byte(xoff) && byte(xoff) < p.curSprs[i].x + 8 {
                    tile := byte(0)
                    cur = p.curSprs[i]
                    pal := (1<<4) | ((cur.attrs & 3) << 2)
                    xsoff := byte(xoff) - cur.x
                    if cur.attrs & (1<<6) != 0 {
                        xsoff = 7-xsoff
                    }
                    ysoff := y-cur.y-1
                    if p.pctrl & (1<<5) != 0 { //8x16
                        if cur.attrs & (1<<7) != 0 {
                            ysoff = 15-ysoff
                        }
                        tile = cur.tile
                        baseSprAddr = word(tile & 1) << 12
                        tile &= ^byte(1)
                        if ysoff > 7 {
                            ysoff -= 8
                            tile |= 1
                        }
                    } else {
                        tile = cur.tile
                        if cur.attrs & (1<<7) != 0 {
                            ysoff = 7-ysoff
                        }
                    }
                    pat := (word(tile) << 4) + baseSprAddr
                    shi := p.getMem(pat+8+word(ysoff))
                    slo := p.getMem(pat+word(ysoff))
                    shi >>= (7-xsoff)
                    shi &= 1
                    shi <<= 1
                    slo >>= (7-xsoff)
                    slo &= 1
                    if (cur.index==0 && (hi|lo) !=0 && (shi|slo) != 0 && bgEnabled) {
                        p.pstat |= 1<<6
                    }
                    if ((hi|lo==0 && shi|slo!=0) || (cur.attrs & (1<<5) == 0)) {
                        if (shi|slo != 0) {
                            coli = 0x3f00 | word(pal | shi | slo)
                            break
                        }
                    }
                }
            }
        }
        color := 0
        if int(p.getMem(coli)) < len(colors) {
            color = colors[p.getMem(coli)]
        }
        p.pixels[int(xoff) + int(y)*256] = color
        //setPixel
        p.fineX++
        p.fineX &= 7
        xoff++
        if p.fineX == 0 {
            if (p.vaddr & 0x1f) == 0x1f {
                p.vaddr ^= 0x400
                p.vaddr -= 0x1f
            } else {
                p.vaddr++
            }
        }
        num--
    }
}

func (p *PPU) drawFrame() {
    p.sl = -2
    moreEvents := true
    for moreEvents {
        select {
        case event := <-sdl.Events:
            switch e := event.(type) {
                case sdl.QuitEvent:
                    sdl.Quit()
                    os.Exit(0)
            }
        default:
            moreEvents = false
        }
    }
    sdl.WM_SetCaption("Hello","")
    p.screen.Flip()
}

func (p *PPU) run() {
    bgEnabled := p.pmask & (1<<3) != 0
    spriteEnabled := p.pmask & (1<<4) != 0
    renderingEnabled := bgEnabled || spriteEnabled
    cycles := int(p.mach.cpu.cycleCount * 3 - p.cycleCount)
    for (p.cycleCount < p.mach.cpu.cycleCount * 3) {
        switch true {
            case p.sl == -2:
                p.doVblank(renderingEnabled)
            case p.sl == -1:
                p.cycleCount += 341
                p.sl += 1
            case p.sl < 240:
                todo := 0
                if 341 - p.cyc > cycles {
                    todo = cycles
                } else {
                    todo = 341 - p.cyc
                }
                y := byte(p.sl)
                if renderingEnabled && p.cyc < 256  {
                    p.renderPixels(byte(p.cyc), y, byte(min(todo, 256-p.cyc)))
                } else if p.cyc >= 257 {}
                p.cyc += todo
                p.cycleCount += uint64(todo)
                if p.cyc == 341 {
                    p.cyc = 0
                    p.sl += 1
                    if renderingEnabled {
                        p.newScanline()
                    }
                }
            case p.sl == 240:
                if 341 - p.cyc > cycles {
                    p.cyc += cycles
                    p.cycleCount += uint64(cycles)
                } else {
                    p.cycleCount += uint64(341 - p.cyc)
                    p.cyc = 0
                    p.sl += 1
                    p.pstat |= (1<<7)
                    p.pstat &= ^byte(1<<6)
                    if p.pctrl & (1<<7) != 0 {
                        p.mach.cpu.nmi()
                    }
                }
            default:
                p.cycleCount += 341 * 20
                p.drawFrame()
        }
    }
}

package gones

import (
    "image"
    "image/png"
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
    mach   *Machine
    cycles chan int
    frames chan []int
    screen []int
    //cycles
    cycleCount uint64
    oddFrame bool
    lastNMI uint64
    vblOff uint64
    NMIOccurred bool
    a12high bool
    //memory
    mem         [0x4000]byte
    memBuf      byte
    mirrorTable [0x4000]word
    latch       bool
    pmask       byte
    pstat       byte
    pctrl       byte
    //prefetch
    bgPrefetch  chan Tile
    bufTile     Tile
    nextSprs         [8]Sprite
    numNextSprs      int
    curSprs          [8]Sprite
    numSprs          int
    //position
    xoff, fineX      byte
    horizScroll      bool
    vertScroll       bool
    currentMirroring int
    vaddr, taddr     word
    nextVaddr        word
    sl               int
    cyc              int
    objMem           [0x100]byte
    objAddr          word
    curTile          Tile
    //debugging
    frameCounter     uint64
    numBytes         uint
    numBytesRead     uint
}

type Sprite struct {
    index             int
    y, tile, attrs, x byte
    patternLo         byte
    patternHi         byte
}

type Tile struct {
    ntAddr word
    ntVal byte
    attr byte
    patternLo byte
    patternHi byte
}

func (s *Sprite) setSpr(index int, m []byte) {
    s.index = index
    s.y, s.tile, s.attrs, s.x = m[0], m[1], m[2], m[3]
}

func makePPU(m *Machine, frames chan []int) *PPU {
    p := PPU{mach: m, frames: frames}
    p.screen = make([]int, 256*240)
    for i := word(0); i < 0x4000; i++ {
        p.mirrorTable[i] = i
        p.mem[i] = 0xff
    }
    for i := int(0); i < 0x100; i++ {
        p.objMem[i] = 0xff
    }
    p.setMirroring(0x3000, 0x2000, 0xf00)
    p.currentMirroring = -1
    p.setNTMirroring(m.rom.mirror)
    p.sl = -2
    p.cycles = make(chan int)
    p.bgPrefetch = make(chan Tile, 128) // not sure about this value
    return &p
}

func (p *PPU) dump() string {
    return fmt.Sprintf("CYC: %d SL: %d VADDR: %4X", p.cyc, p.sl, p.vaddr)
}

func (p *PPU) setMirroring(from word, to word, n word) {
    for i := word(0); i < n; i++ {
        p.mirrorTable[from+i] = to + i
    }
}

func (p *PPU) setNTMirroring(t int) {
    if t == p.currentMirroring {
        return
    }
    p.currentMirroring = t
    switch t {
    case VERTICAL:
        p.setMirroring(0x2000, 0x2000, 0x400)
        p.setMirroring(0x2400, 0x2400, 0x400)
        p.setMirroring(0x2800, 0x2000, 0x400)
        p.setMirroring(0x2c00, 0x2400, 0x400)
    case HORIZONTAL:
        p.setMirroring(0x2000, 0x2000, 0x400)
        p.setMirroring(0x2400, 0x2000, 0x400)
        p.setMirroring(0x2800, 0x2400, 0x400)
        p.setMirroring(0x2c00, 0x2400, 0x400)
    case SINGLE_LOWER:
        p.setMirroring(0x2000, 0x2000, 0x400)
        p.setMirroring(0x2400, 0x2000, 0x400)
        p.setMirroring(0x2800, 0x2000, 0x400)
        p.setMirroring(0x2c00, 0x2000, 0x400)
    case SINGLE_UPPER:
        p.setMirroring(0x2000, 0x2400, 0x400)
        p.setMirroring(0x2400, 0x2400, 0x400)
        p.setMirroring(0x2800, 0x2400, 0x400)
        p.setMirroring(0x2c00, 0x2400, 0x400)
    default:
        break
    }
}

func (p *PPU) readRegister(num int) byte {
    ret := byte(0)
    switch num {
    case 2:
        ret = p.pstat
        p.pstat &= ^byte(1 << 7)
        p.latch = false
        cycles := p.cycleCount
        if cycles - p.lastNMI < 3 {
            p.mach.suppressNMI()
            if cycles - p.lastNMI == 0 {
                ret = p.pstat
            }
        }
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
        if p.pctrl&(1<<2) != 0 {
            p.vaddr += 32
        } else {
            p.vaddr += 1
        }
        p.vaddr &= 0x3fff
        p.a12high = p.vaddr & 0x1000 != 0
        return ret
    }
    return 0
}

func (p *PPU) writeRegister(num int, val byte) {
    switch num {
    case 0:
        p.pctrl = val
        cycles := p.mach.cpu.cycleCount*3
        if p.pctrl & (1<<7) != 0 {
            if (p.pstat & (1<<7) != 0 || (cycles - p.vblOff <= 2)) && !p.NMIOccurred {
                p.mach.requestNMI()
                p.NMIOccurred = true
            }
        } else {
            p.NMIOccurred = false
            if (cycles - p.lastNMI < 6) {
                p.mach.suppressNMI()
            }
        }
        p.taddr &= (^word(0x3 << 10))
        p.taddr |= word(val&0x3) << 10
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
            p.taddr |= word(val>>3) << 5
            p.taddr |= word(val&0x7) << 12
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
            p.a12high = p.vaddr & 0x1000 != 0
            if p.cyc >= 251 && p.sl < 240 && p.sl >= 0 {
                p.vertScroll = true
            }
        } else {
            p.taddr &= 0xff
            p.taddr |= word(val&0x3f) << 8
        }
        p.latch = !p.latch
    case 7:
        p.setMem(p.vaddr, val)
        if p.pctrl&(1<<2) != 0 {
            p.vaddr += 32
        } else {
            p.vaddr += 1
        }
        p.vaddr &= 0x3fff
        p.a12high = p.vaddr & 0x1000 != 0
    }
}

func (p *PPU) getMem(addr word) byte {
    switch true {
    case addr < 0x2000:
        p.a12high = addr & 0x1000 != 0
        chr_bank := (addr&p.mach.rom.chr_bank_mask)>>p.mach.rom.chr_bank_shift
        return p.mach.rom.chr_rom[chr_bank][addr&(^p.mach.rom.chr_bank_mask)]
    case addr < 0x3000:
        return p.mem[p.mirrorTable[addr]]
    case addr < 0x3f00:
        return p.getMem(addr - 0x1000)
    default:
        if addr&0xf == 0 {
            addr = 0
        }
        return p.mem[0x3f00+(addr&0x1f)]
    }
    return 0
}

func (p *PPU) setMem(addr word, val byte) {
    switch true {
    case addr < 0x2000:
        p.a12high = addr & 0x1000 != 0
        chr_bank := (addr&p.mach.rom.chr_bank_mask)>>p.mach.rom.chr_bank_shift
        p.mach.rom.chr_rom[chr_bank][addr&(^p.mach.rom.chr_bank_mask)] = val
    case addr < 0x3f00:
        p.mem[p.mirrorTable[addr]] = val
    default:
        if addr&0xf == 0 {
            addr = 0
        }
        p.mem[0x3f00+(addr&0x1f)] = val
    }
}

func (p *PPU) newScanline() {
    p.vertScroll = false
    p.horizScroll = false
    p.fineX = p.xoff
    p.numBytesRead = 0
    for len(p.bgPrefetch) > 2 {
        <-p.bgPrefetch
    }
    p.curTile = <-p.bgPrefetch
    //sprites
    p.numSprs = p.numNextSprs
    for i := 0; i < p.numSprs; i++ {
        p.curSprs[i] = p.nextSprs[i]
    }

    p.numNextSprs = 0
    curY := p.sl
    if curY == 240 {
        p.numNextSprs = 0
        return
    }
    s := Sprite{}
    for i := 0; i < 64; i++ {
        (&s).setSpr(i, p.objMem[i*4:i*4+4])
        if int(s.y) <= curY && (curY < int(s.y)+8 || (p.pctrl&(1<<5) != 0 && curY < int(s.y)+16)) {
            if p.numNextSprs == 8 {
                break
            }
            p.nextSprs[p.numNextSprs] = s
            p.numNextSprs++
        }
    }
}

func (p *PPU) updateVertScroll() {
    fineY := (p.vaddr & 0x7000) >> 12
    if fineY == 7 {
        if p.vaddr&0x3ff >= 0x3e0 {
            p.vaddr &= ^word(0x3ff)
        } else {
            p.vaddr += 0x20
            if p.vaddr&0x3ff >= 0x3c0 {
                p.vaddr &= ^word(0x3ff)
                p.vaddr ^= 0x800
            }
        }
    }
    p.vaddr &= ^word(0x7000)
    p.vaddr |= (fineY + 1) & 7 << 12
}

func (p *PPU) doVblank(renderingEnabled bool) {
    cycles := int(p.mach.cpu.cycleCount*3 - p.cycleCount)
    if 341-p.cyc > cycles {
        p.cyc += cycles
        p.cycleCount += uint64(cycles)
    } else {
        p.cycleCount += uint64(341 - p.cyc)
        p.cyc = 0
        p.sl += 1
        if renderingEnabled {
            p.fineX = p.xoff
        }
    }
}

func (p *PPU) prefetchBytes(start int, cycles int) {
    bgEnabled := p.pmask&(1<<3) != 0
    spriteEnabled := p.pmask&(1<<4) != 0
    if !bgEnabled && !spriteEnabled {
        return
    }
    basePtAddr := word(0x0)
    if p.pctrl&(1<<4) != 0 {
        basePtAddr = 0x1000
    }
    baseSprAddr := word(0x0)
    if p.pctrl&(1<<3) != 0 {
        baseSprAddr = 0x1000
    }
    for j := start; j < (start+cycles); j++ {
        if j & 1 == 0 { continue }
        i := j/2
        switch true {
        case i < 124:
            //load bg tile data
            fineY := (p.vaddr >> 12) & 7
            ntaddr := 0x2000 + (p.vaddr & 0xfff)
            p.bufTile.ntAddr = ntaddr
            switch i%4 {
            case 0:
                //fetch nt byte
                p.bufTile.ntVal = p.getMem(ntaddr)
            case 1:
                atBase := (ntaddr & ^word(0x3ff)) + 0x3c0
                p.bufTile.attr = p.getMem(atBase + ((ntaddr & 0x1f) >>2) + ((ntaddr & 0x3e0)>>7)*8)
            case 2:
                ptAddr := (word(p.bufTile.ntVal) << 4) + basePtAddr;
                p.bufTile.patternLo = p.getMem(ptAddr + fineY)
            case 3:
                ptAddr := (word(p.bufTile.ntVal) << 4) + basePtAddr;
                p.bufTile.patternHi = p.getMem(ptAddr + 8 + fineY)
                p.bgPrefetch <- p.bufTile
                p.numBytes++
                if (p.vaddr & 0x1f) == 0x1f {
                    p.vaddr ^= 0x400
                    p.vaddr -= 0x1f
                } else {
                    p.vaddr++
                }
            }
//        case 124 <= i && i < 128:
        case 128 <= i && i < 160:
            //sprite pattern fetches for next sl
            cur := p.nextSprs[(i-128)/4]
            tile := byte(0)
            ysoff := byte(p.sl) - cur.y
            if p.pctrl&(1<<5) != 0 { //8x16
                if cur.attrs&(1<<7) != 0 {
                    ysoff = 15 - ysoff
                }
                tile = cur.tile
                baseSprAddr = word(tile&1) << 12
                tile &= ^byte(1)
                if ysoff > 7 {
                    ysoff -= 8
                    tile |= 1
                }
            } else {
                tile = cur.tile
                if cur.attrs&(1<<7) != 0 {
                    ysoff = 7 - ysoff
                }
            }
            pat := (word(tile) << 4) + baseSprAddr
            switch i%4 {
            case 0:
                break
            case 2:
                p.nextSprs[(i-128)/4].patternLo = p.getMem(pat + word(ysoff))
            case 3:
                p.nextSprs[(i-128)/4].patternHi = p.getMem(pat + 8 + word(ysoff))
            }
        case 160 <= i && i < 168:
            //for next scanline TODO repeated and doesn't work
            fineY := (p.vaddr >> 12) & 7
            ntaddr := 0x2000 + (p.vaddr & 0xfff)
            p.bufTile.ntAddr = ntaddr
            switch i%4 {
            case 0:
                //fetch nt byte
                p.bufTile.ntVal = p.getMem(ntaddr)
            case 1:
                atBase := (ntaddr & ^word(0x3ff)) + 0x3c0
                p.bufTile.attr = p.getMem(atBase + ((ntaddr & 0x1f) >>2) + ((ntaddr & 0x3e0)>>7)*8)
            case 2:
                ptAddr := (word(p.bufTile.ntVal) << 4) + basePtAddr;
                p.bufTile.patternLo = p.getMem(ptAddr + fineY)
            case 3:
                ptAddr := (word(p.bufTile.ntVal) << 4) + basePtAddr;
                p.bufTile.patternHi = p.getMem(ptAddr + 8 + fineY)
                p.bgPrefetch <- p.bufTile
                p.numBytes++
                if (p.vaddr & 0x1f) == 0x1f {
                    p.vaddr ^= 0x400
                    p.vaddr -= 0x1f
                } else {
                    p.vaddr++
                }
            }
        }
        if i == 68 {
            if p.numNextSprs == 8 {
                p.pstat |= 1<<5
            }
        } else if i == 126 {
            p.updateVertScroll()
        } else if i == 128 {
            p.vaddr &= ^word(0x041f)
            p.vaddr |= p.taddr & 0x1f
            p.vaddr |= p.taddr & 0x400
            //p.fineX = p.xoff TODO
        }
    }
}

func (p *PPU) renderPixels(x byte, y byte, num byte) {
    bgEnabled := p.pmask&(1<<3) != 0
    spriteEnabled := p.pmask&(1<<4) != 0
    xoff := p.cyc
    for num != 0 {
        row := (p.curTile.ntAddr >> 6) & 1
        col := (p.curTile.ntAddr & 2) >> 1
        atVal := p.curTile.attr
        atVal >>= 4*row + 2*col
        atVal &= 3
        atVal <<= 2
        hi := p.curTile.patternHi
        lo := p.curTile.patternLo
        hi >>= (7 - p.fineX)
        hi &= 1
        hi <<= 1
        lo >>= (7 - p.fineX)
        lo &= 1
        coli := word(0x3f00)
        if hi|lo != 0 && bgEnabled && !(xoff < 8 && (p.pmask&2 == 0)) {
            coli |= word(atVal | hi | lo)
        }
        //TODO sprites
        if spriteEnabled && !(xoff < 8 && (p.pmask&4 == 0)) {
            cur := Sprite{}
            for i := 0; i < p.numSprs; i++ {
                if int(p.curSprs[i].x) <= xoff && xoff < int(p.curSprs[i].x)+8 {
                    cur = p.curSprs[i]
                    pal := (1 << 4) | ((cur.attrs & 3) << 2)
                    xsoff := byte(xoff) - cur.x
                    if cur.attrs&(1<<6) != 0 {
                        xsoff = 7 - xsoff
                    }
                    shi := cur.patternHi
                    slo := cur.patternLo
                    shi >>= (7 - xsoff)
                    shi &= 1
                    shi <<= 1
                    slo >>= (7 - xsoff)
                    slo &= 1
                    if cur.index == 0 && (hi|lo) != 0 && (shi|slo) != 0 && bgEnabled && !(xoff < 8 && (p.pmask & 2 == 0)) && xoff < 255 {
                        p.pstat |= 1 << 6
                    }
                    if (hi|lo == 0) || (cur.attrs&(1<<5) == 0) {
                        if shi|slo != 0 {
                            coli = 0x3f00 | word(pal|shi|slo)
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
        p.screen[int(y)*256+int(xoff)] = color
        p.fineX++
        p.fineX &= 7
        xoff++
        num--
        if p.fineX == 0 {
            if len(p.bgPrefetch) < 1 {
            } else {
                p.curTile = <-p.bgPrefetch
            }
            p.numBytesRead++
        }
    }
}

func (p *PPU) drawFrame() {
    p.sl = -2
    p.frames <- p.screen
}

func (p *PPU) run() {
    bgEnabled := p.pmask&(1<<3) != 0
    spriteEnabled := p.pmask&(1<<4) != 0
    renderingEnabled := bgEnabled || spriteEnabled
    for p.cycleCount < p.mach.cpu.cycleCount*3 {
        cycles := int(p.mach.cpu.cycleCount*3 - p.cycleCount)
        switch true {
        case p.sl == -2:
            //fmt.Println(len(p.bgPrefetch))
            p.doVblank(renderingEnabled)
        case p.sl == -1:
            switch p.cyc {
            case 0:
                p.pstat &= ^byte(1 << 7)
                p.pstat &= ^byte(1 << 6)
                p.pstat &= ^byte(1 << 5)
                p.vblOff = p.cycleCount
                p.cycleCount += 4
                p.cyc += 4
            case 4:
                p.cycleCount += 256
                p.cyc += 256
                //p.prefetchBytes(0, 4)
            case 260:
                p.cycleCount += 44
                p.cyc += 44
                //p.prefetchBytes(4, 256)
            case 304:
                if bgEnabled {
                    //p.prefetchBytes(260, 44) //TODO inaccurate
                    p.vaddr = p.taddr
                }
                p.cycleCount += 36
                p.cyc += 36
            case 340:
                if bgEnabled{
                    if p.oddFrame {
                        p.cycleCount -= 1
                    }
                    //p.prefetchBytes(304, 37) //TODO inaccurate
                }
                p.oddFrame = !p.oddFrame
                p.cycleCount++
                p.cyc++
            case 341:
                if bgEnabled {
                    p.prefetchBytes(320, 21) //TODO inaccurate
                    //p.bgPrefetch <- Tile{}
                    //p.bgPrefetch <- Tile{}
                }
                p.cyc = 0
                p.sl += 1
                if bgEnabled {
                    p.newScanline()
                }
                p.numBytesRead++
            }
        case p.sl < 240:
            todo := 0
            if 341-p.cyc > cycles {
                todo = cycles
            } else {
                todo = 341 - p.cyc
            }
            y := byte(p.sl)
            if renderingEnabled {
                p.prefetchBytes(p.cyc, todo)
                if p.cyc < 256 {
                    p.renderPixels(byte(p.cyc), y, byte(min(todo, 256-p.cyc)))
                }
            }
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
            if 341-p.cyc > cycles {
                p.cyc += cycles
                p.cycleCount += uint64(cycles)
            } else {
                p.cycleCount += uint64(341 - p.cyc)
                p.cyc = 0
                p.sl += 1
                p.lastNMI = p.cycleCount
                p.pstat |= (1 << 7)
                if p.pctrl&(1<<7) != 0 {
                    p.mach.requestNMI()
                    p.NMIOccurred = true
                } else {
                    p.NMIOccurred = false
                }
            }
        case p.sl == 241:
            if 341-p.cyc > cycles {
                p.cycleCount += uint64(cycles)
                p.cyc += cycles
            } else {
                p.cycleCount += uint64(341 - p.cyc)
                p.cyc = 0
                p.sl += 1
            }
        default:
            //vblank for 18 + 2 = 20 scanlines
            p.cycleCount += 341 * 18
            p.drawFrame()
        }
    }
}

func (p *PPU) dumpNTs() {
    img := image.NewRGBA(512, 480)
    basePtAddr := word(0)
    if p.pctrl&(1<<4) != 0 {
        basePtAddr = 0x1000
    }
    x := uint(0)
    y := uint(0)
    for nt := word(0x2000); nt < 0x3000; nt += 0x400 {
        at_base := nt + 0x3c0
        for ntaddr := nt; ntaddr < nt+0x3c0; ntaddr++ {
            ntval := p.getMem(ntaddr)
            ptaddr := (word(ntval) << 4) + basePtAddr
            row := (ntaddr >> 6) & 1
            col := (ntaddr & 2) >> 1
            atval := p.getMem(at_base + ((ntaddr & 0x1f) >> 2) + ((ntaddr&0x3e0)>>7)*8)
            atval >>= 4*row + 2*col
            atval &= 3
            atval <<= 2
            for fy := uint(0); fy < 8; fy++ {
                for fx := uint(0); fx < 8; fx++ {
                    hi := p.getMem(ptaddr + word(8+fy))
                    lo := p.getMem(ptaddr + word(fy))
                    hi >>= (7 - fx)
                    hi &= 1
                    hi <<= 1
                    lo >>= 7 - fx
                    lo &= 1
                    coli := word(0x3f00)
                    if hi|lo != 0 {
                        coli |= word(atval | hi | lo)
                    }
                    color := colors[p.getMem(coli)]
                    img.Set(int(x+fx), int(y+fy), intToColor(color))
                }
            }
            x += 8
            if x%256 == 0 {
                x -= 256
                y += 8
            }
        }
        x += 256
        y -= 240
        if x == 512 {
            x = 0
            y = 240
        }
    }
    f, e := os.Create("nt.png")
    if f == nil {
        fmt.Printf("error opening file. %v\n", e.String())
    }
    png.Encode(f, img)
}

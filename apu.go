package gones

import "fmt"

type Pulse struct {
    dutyCycle byte
    lengthHalt bool
    envelope byte
    timer word
    lengthEnabled bool
    lengthCounter byte
}

func (p *Pulse) enableLength(en bool) {
    p.lengthEnabled = en
    if !p.lengthEnabled {
        p.lengthCounter = 0
    }
}

func (p *Pulse) clockLengthCounter() {
    if !p.lengthEnabled { return }
    if p.lengthCounter > 0 {
        p.lengthCounter -= 1
    }
    if p.lengthCounter == 0 {
        if p.lengthHalt {
            //do something
        }
    }
}

func (p *Pulse) lengthNonzero() bool {
    return p.lengthCounter > 0
}

func (p *Pulse) writeRegister(num byte, val byte) {
    switch num % 4 {
    case 0:
        p.dutyCycle = (val & 0xc0) >> 6
        p.lengthHalt = val & 0x20 != 0
        p.envelope = val & 0x1f
    case 1:
        //sweep
        break
    case 2:
        p.timer &= ^word(0xf)
        p.timer |= word(val)
    case 3:
        p.lengthCounter = lengthTable[(val & 0xf8)>>3]
        p.timer &= ^word(0x70)
        p.timer |= word(val & 0x7) << 8
    }
}

func (p *Pulse) readRegister(num byte) byte {
    return 0
}

type APU struct {
    m *Machine
    //registers
    status byte
    //channels
    p1, p2 Pulse
    //triangle, noise, dmc
    //frame counter
    frameMode bool
    oddClock bool
    frameIrq bool
    frameInterrupt bool
    frameCycles int
    sequencerStatus byte
    counter int
}

func makeAPU(mach *Machine) *APU {
    a := APU{}
    a.m = mach
    return &a
}

func (a *APU) clockSequencer() {
    if a.frameMode {
        switch a.sequencerStatus {
        case 1, 3:
            //clock both
            break
        case 2, 4:
            //clock one
            break
        }
        a.sequencerStatus = (a.sequencerStatus + 1) % 5
    } else {
        switch a.sequencerStatus {
        case 0, 2:
            break
        case 1:
            a.p1.clockLengthCounter()
            a.p2.clockLengthCounter()
        case 3:
            if a.frameIrq {
                a.frameInterrupt = true
            }
            a.p1.clockLengthCounter()
            a.p2.clockLengthCounter()
        }
        a.sequencerStatus = (a.sequencerStatus + 1) % 4
    }
}

func (a *APU) writeRegister(num byte, val byte) {
    switch num {
    case 0x0,0x1,0x2,0x3:
        a.p1.writeRegister(num, val)
    case 0x4,0x5,0x6,0x7:
        a.p2.writeRegister(num, val)
    case 0x15:
        a.p1.enableLength(val & 0x1 != 0)
        a.p2.enableLength(val & 0x2 != 0)
    case 0x17:
        a.frameMode = val & 0x80 != 0
        if val & 0x40 != 0 {
            a.frameInterrupt = false
            a.frameIrq = false
        } else {
            a.frameIrq = true
        }
    default:
        break
        fmt.Printf("weird APU register %v\n", num)
    }
}

func (a *APU) readRegister(num byte) byte {
    oldStatus := byte(0)
    switch num {
    case 0x15:
        if a.frameInterrupt {
            oldStatus |= 1<<6
        }
        if a.p1.lengthNonzero() {
            oldStatus |= 1
        }
        if a.p2.lengthNonzero() {
            oldStatus |= 2
        }
        a.frameInterrupt = false
        return oldStatus
    case 0x17:
        return 0
    default:
        return 0
    }
    return 0
}

func (a *APU) update(cycles int) {
    a.frameCycles += cycles
    if a.p1.lengthNonzero() {
        a.counter += cycles
    }
    if (a.oddClock && a.frameCycles > 7457) || a.frameCycles > 7458 {
        if !a.oddClock {
            a.frameCycles -= 1
        }
        a.frameCycles -= 7457
        a.clockSequencer()
        a.oddClock = !a.oddClock
    }
    if a.frameInterrupt {
        //a.m.requestIrq()
    }
}

var lengthTable = [0x20]byte{ 10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14, 12, 15, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30 }

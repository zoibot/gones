package gones

import "fmt"
import "os"

type Mapper interface {
    load(rom *ROM)
    prgWrite(addr word, val byte)
    update(m *Machine)
    name() string
}

func loadMapper(num byte, rom *ROM) Mapper {
    var m Mapper
    switch num {
    case 0:
        m = new(NROM)
    case 1:
        m = new(MMC1)
    case 2:
        m = new(UNROM)
    case 3:
        m = new(CNROM)
    case 4:
        m = new(MMC3)
    case 7:
        m = new (AXROM)
    default:
        fmt.Printf("Unsupported Mapper: %d\n", num)
        os.Exit(1)
    }
    m.load(rom)
    fmt.Printf("Mapper: %d %s\n", num, m.name())
    return m
}

type NROM struct{}

func (n *NROM) load(rom *ROM) {
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000*int(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
    rom.chr_bank_mask = 0x1000
    rom.chr_bank_shift = 12
    rom.prg_bank_mask = 0x4000
    rom.prg_bank_shift = 14
}

func (n *NROM) prgWrite(addr word, val byte) {}

func (n *NROM) update(m *Machine) {}

func (n *NROM) name() string { return "NROM" }

type MMC1 struct {
    control, loadr, shift, prg_bank byte
    rom                             *ROM
}

func (m *MMC1) load(rom *ROM) {
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000*int(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
    rom.chr_bank_mask = 0x1000
    rom.chr_bank_shift = 12
    rom.prg_bank_mask = 0x4000
    rom.prg_bank_shift = 14
    m.control = 0xc
    m.shift = 0
    m.loadr = 0
    m.prg_bank = 0
    m.rom = rom
}

func (m *MMC1) prgWrite(addr word, val byte) {
    m.loadr |= (val & 1) << m.shift
    m.shift++
    if val&0x80 != 0 {
        m.loadr = 0
        m.shift = 0
        m.control |= 0xc
        return
    }
    if m.shift == 5 {
        if addr < 0xa000 {
            switch m.loadr & 3 {
            case 0:
                m.rom.mirror = SINGLE_LOWER
            case 1:
                m.rom.mirror = SINGLE_UPPER
            case 2:
                m.rom.mirror = VERTICAL
            case 3:
                m.rom.mirror = HORIZONTAL
            }
            if m.control&0xc != m.loadr&0xc {
                m.control = m.loadr
                m.updatePrgBanks()
            }
            m.control = m.loadr
        } else if addr < 0xc000 {
            if m.control&(1<<4) != 0 {
                //4kb mode
                m.rom.chr_rom[0] = m.rom.chr_banks[0x1000*int(m.loadr):]
            } else {
                m.rom.chr_rom[0] = m.rom.chr_banks[0x1000*int(m.loadr&0x1e):]
                m.rom.chr_rom[1] = m.rom.chr_banks[0x1000*int(m.loadr|1):]
            }
        } else if addr < 0xe000 {
            if m.control&(1<<4) != 0 {
                //4kb mode
                m.rom.chr_rom[1] = m.rom.chr_banks[0x1000*int(m.loadr):]
            }
        } else {
            m.prg_bank = m.loadr
            m.updatePrgBanks()
        }
        m.shift = 0
        m.loadr = 0
    }
}

func (m *MMC1) update(mach *Machine) {}

func (m *MMC1) updatePrgBanks() {
    switch m.control & 0xc {
    case 0, 4:
        m.rom.prg_rom[0] = m.rom.prg_banks[0x4000*int(m.prg_bank&0xe):]
        m.rom.prg_rom[1] = m.rom.prg_banks[0x4000*int(m.prg_bank|1):]
    case 8:
        m.rom.prg_rom[0] = m.rom.prg_banks
        m.rom.prg_rom[1] = m.rom.prg_banks[0x4000*int(m.prg_bank):]
    case 0xc:
        m.rom.prg_rom[0] = m.rom.prg_banks[0x4000*int(m.prg_bank):]
        m.rom.prg_rom[1] = m.rom.prg_banks[0x4000*int(m.rom.prg_size-1):]
    }
}

func (m *MMC1) name() string { return "MMC1" }

type UNROM struct {
    rom *ROM
}

func (u *UNROM) load(rom *ROM) {
    u.rom = rom
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000*int(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
    rom.chr_bank_mask = 0x1000
    rom.chr_bank_shift = 12
    rom.prg_bank_mask = 0x4000
    rom.prg_bank_shift = 14
}

func (u *UNROM) prgWrite(addr word, val byte) {
    bank := int(val & 7)
    u.rom.prg_rom[0] = u.rom.prg_banks[0x4000*bank:]
}

func (u *UNROM) update(m *Machine) {}

func (u *UNROM) name() string {
    return "UNROM"
}

type CNROM struct {
    rom *ROM
}

func (c *CNROM) load(rom *ROM) {
    c.rom = rom
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000*int(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
    rom.chr_bank_mask = 0x1000
    rom.chr_bank_shift = 12
    rom.prg_bank_mask = 0x4000
    rom.prg_bank_shift = 14
}

func (c *CNROM) prgWrite(addr word, val byte) {
    bank := int(val & 3)
    c.rom.chr_rom[0] = c.rom.chr_banks[0x2000*bank:]
    c.rom.chr_rom[1] = c.rom.chr_banks[0x2000*bank+0x1000:]
}

func (c *CNROM) update(m *Machine) {}

func (c *CNROM) name() string {
    return "CNROM"
}

type MMC3 struct {
    rom *ROM
    bankSelect byte
    currentChrBanks [6]int
    currentPrgBanks [3]int
    bankConfiguration byte
    irqLatch byte
    irqEnabled bool
    irqWaiting bool
    irqCounter byte
    a12high bool
}

func (c *MMC3) load(rom *ROM) {
    c.rom = rom
    rom.prg_rom[3] = rom.prg_banks[0x2000*int(rom.prg_size*2-1):]
    for i := 0; i < 6; i++ {
        c.currentChrBanks[i] = i
    }
    c.currentPrgBanks[0] = 0
    c.currentPrgBanks[1] = 1
    c.updateChrBanks()
    c.updatePrgBanks()
    rom.chr_bank_mask = 0xfc00
    rom.chr_bank_shift = 10
    rom.prg_bank_mask = 0x6000
    rom.prg_bank_shift = 13
}

func (c *MMC3) prgWrite(addr word, val byte) {
    switch addr & 1 {
    case 0:
        switch true {
            case addr < 0xa000:
                //bank select
                c.bankSelect = val & 7
                c.bankConfiguration = (val & 0xc0) >> 6
            case addr < 0xc000:
                //mirroring
                if val & 1 == 0 {
                    c.rom.mirror = VERTICAL
                } else {
                    c.rom.mirror = HORIZONTAL
                }
            case addr < 0xe000:
                //irq latch
                c.irqLatch = val
            default:
                //irq disable
                c.irqEnabled = false
                //acknowledge waiting
                c.irqWaiting = false
        }
    case 1:
        switch true {
            case addr < 0xa000:
                //set bank
                v := int(val)
                switch c.bankSelect {
                    case 0:
                        c.currentChrBanks[0] = v & 0xfe
                    case 1:
                        c.currentChrBanks[1] = v & 0xfe
                    case 2:
                        c.currentChrBanks[2] = v
                    case 3:
                        c.currentChrBanks[3] = v
                    case 4:
                        c.currentChrBanks[4] = v
                    case 5:
                        c.currentChrBanks[5] = v
                    case 6:
                        c.currentPrgBanks[0] = v
                    case 7:
                        c.currentPrgBanks[1] = v
                }
                c.updatePrgBanks()
                c.updateChrBanks()
            case addr < 0xc000:
                //prg ram
            case addr < 0xe000:
                //irq reload
                c.irqCounter = 0
            default:
                //irq enable
                c.irqEnabled = true
        }
    }
}

func (c *MMC3) updateChrBanks() {
    if c.bankConfiguration & 2 == 0 {
        //two four
        c.rom.chr_rom[0] = c.rom.chr_banks[c.currentChrBanks[0]*0x400:]
        c.rom.chr_rom[1] = c.rom.chr_banks[c.currentChrBanks[0]*0x400 + 0x400:]
        c.rom.chr_rom[2] = c.rom.chr_banks[c.currentChrBanks[1]*0x400:]
        c.rom.chr_rom[3] = c.rom.chr_banks[c.currentChrBanks[1]*0x400 + 0x400:]
        c.rom.chr_rom[4] = c.rom.chr_banks[c.currentChrBanks[2]*0x400:]
        c.rom.chr_rom[5] = c.rom.chr_banks[c.currentChrBanks[3]*0x400:]
        c.rom.chr_rom[6] = c.rom.chr_banks[c.currentChrBanks[4]*0x400:]
        c.rom.chr_rom[7] = c.rom.chr_banks[c.currentChrBanks[5]*0x400:]
    } else {
        //four two
        c.rom.chr_rom[0] = c.rom.chr_banks[c.currentChrBanks[2]*0x400:]
        c.rom.chr_rom[1] = c.rom.chr_banks[c.currentChrBanks[3]*0x400:]
        c.rom.chr_rom[2] = c.rom.chr_banks[c.currentChrBanks[4]*0x400:]
        c.rom.chr_rom[3] = c.rom.chr_banks[c.currentChrBanks[5]*0x400:]
        c.rom.chr_rom[4] = c.rom.chr_banks[c.currentChrBanks[0]*0x400:]
        c.rom.chr_rom[5] = c.rom.chr_banks[c.currentChrBanks[0]*0x400 + 0x400:]
        c.rom.chr_rom[6] = c.rom.chr_banks[c.currentChrBanks[1]*0x400:]
        c.rom.chr_rom[7] = c.rom.chr_banks[c.currentChrBanks[1]*0x400 + 0x400:]
    }
}

func (c *MMC3) updatePrgBanks() {
    if c.bankConfiguration & 1 == 0 {
        c.rom.prg_rom[0] = c.rom.prg_banks[0x2000 * c.currentPrgBanks[0]:]
        c.rom.prg_rom[1] = c.rom.prg_banks[0x2000 * c.currentPrgBanks[1]:]
        c.rom.prg_rom[2] = c.rom.prg_banks[0x2000 * int(c.rom.prg_size*2-2):]
    } else {
        c.rom.prg_rom[0] = c.rom.prg_banks[0x2000 * int(c.rom.prg_size*2-2):]
        c.rom.prg_rom[1] = c.rom.prg_banks[0x2000 * c.currentPrgBanks[1]:]
        c.rom.prg_rom[2] = c.rom.prg_banks[0x2000 * c.currentPrgBanks[0]:]
    }
}

func (c *MMC3) update(m *Machine) {
    if !c.a12high && m.ppu.a12high {
        c.clockCounter()
    }
    c.a12high = m.ppu.a12high
    if c.irqWaiting && c.irqEnabled {
        m.requestIrq()
    }
}

func (c *MMC3) clockCounter() {
    fmt.Printf("clocking counter %v\n", c.irqCounter)
    if c.irqCounter > 0 {
        c.irqCounter--
    } else {
        c.irqCounter = c.irqLatch
    }
    if c.irqCounter == 0 && c.irqEnabled {
        c.irqWaiting = true
    }
    fmt.Printf("clocking counter now %v\n", c.irqCounter)
}

func (c *MMC3) name() string {
    return "MMC3"
}

type AXROM struct{
    rom *ROM
}

func (a *AXROM) load(rom *ROM) {
    a.rom = rom
    a.rom.prg_rom[0] = a.rom.prg_banks
    a.rom.prg_rom[1] = a.rom.prg_banks[0x4000:]//*int(a.rom.prg_size-1):]
    a.rom.chr_rom[0] = a.rom.chr_banks
    a.rom.chr_rom[1] = a.rom.chr_banks[0x1000:]
    rom.chr_bank_mask = 0x1000
    rom.chr_bank_shift = 12
    rom.prg_bank_mask = 0x4000
    rom.prg_bank_shift = 14
}

func (a *AXROM) prgWrite(addr word, val byte) {
    a.rom.prg_rom[0] = a.rom.prg_banks[0x8000 * int(val&7):]
    a.rom.prg_rom[1] = a.rom.prg_banks[0x8000 * int(val&7) + 0x4000:]
    if val & 0x10 != 0 {
        a.rom.mirror = SINGLE_LOWER
    } else {
        a.rom.mirror = SINGLE_UPPER
    }
}

func (a *AXROM) update(m *Machine) {}

func (a *AXROM) name() string { return "AXROM" }


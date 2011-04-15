package gones

import "fmt"

type Mapper interface {
    load(rom *ROM)
    prgWrite(addr word, val byte)
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
    }
    m.load(rom)
    fmt.Printf("Mapper: %d %s\n", num, m.name())
    return m
}

type NROM struct {}

func (n *NROM) load(rom *ROM) {
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000 * word(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
}

func (n *NROM) prgWrite(addr word, val byte) {}

func (n *NROM) name() string { return "NROM" }

type MMC1 struct {
    control, loadr, shift, prg_bank byte
    rom *ROM
}

func (m *MMC1) load(rom *ROM) {
    rom.prg_rom[0] = rom.prg_banks
    rom.prg_rom[1] = rom.prg_banks[0x4000 * word(rom.prg_size-1):]
    rom.chr_rom[0] = rom.chr_banks
    rom.chr_rom[1] = rom.chr_banks[0x1000:]
    m.control = 0xc
    m.shift = 0
    m.loadr = 0
    m.prg_bank = 0
    m.rom = rom
}

func (m *MMC1) prgWrite(addr word, val byte) {
    m.loadr |= (val & 1) << m.shift
    m.shift++
    if val & 0x80 != 0 {
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
            if m.control & 0xc != m.loadr & 0xc {
                m.control = m.loadr
                m.updatePrgBanks()
            }
            m.control = m.loadr
        } else if addr < 0xc000 {
            if m.control & (1<<5) != 0 {
                //4kb mode
                m.rom.chr_rom[0] = m.rom.chr_banks[0x1000*int(m.loadr):]
            } else {
                m.rom.chr_rom[0] = m.rom.chr_banks[0x1000*int(m.loadr&0x1e):]
                m.rom.chr_rom[1] = m.rom.chr_banks[0x1000*int(m.loadr|1):]
            }
        } else if addr < 0xe000 {
            if m.control & (1<<5) != 0 {
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

func (m *MMC1) updatePrgBanks() {
    switch m.control & 0xc {
        case 0, 4:
            m.rom.prg_rom[0] = m.rom.prg_banks[0x4000 * int(m.prg_bank & 0xe):]
            m.rom.prg_rom[1] = m.rom.prg_banks[0x4000 * int(m.prg_bank | 1):]
        case 8:
            m.rom.prg_rom[0] = m.rom.prg_banks
            m.rom.prg_rom[1] = m.rom.prg_banks[0x4000 * int(m.prg_bank):]
        case 0xc:
            m.rom.prg_rom[0] = m.rom.prg_banks[0x4000 * int(m.prg_bank):]
            m.rom.prg_rom[1] = m.rom.prg_banks[0x4000 * int(m.rom.prg_size-1):]
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
}

func (u *UNROM) prgWrite(addr word, val byte) {
    bank := int(val & 7)
    //fmt.Printf("switching bank: %v %04x %04x\n", bank, 0x4000*bank, len(u.rom.prg_banks))
    u.rom.prg_rom[0] = u.rom.prg_banks[0x4000 * bank:]
}

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
}

func (c *CNROM) prgWrite(addr word, val byte) {
    bank := int(val & 3)
    c.rom.chr_rom[0] = c.rom.chr_banks[0x2000 * bank:]
    c.rom.chr_rom[1] = c.rom.chr_banks[0x2000 * bank + 0x1000:]
}

func (c *CNROM) name() string {
    return "CNROM"
}

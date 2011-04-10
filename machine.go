package gones


type Machine struct {
    cpu *CPU
    ppu *PPU
    rom *ROM
}

func MakeMachine(romname string) *Machine {
    m := &Machine{}
    m.cpu = makeCPU(m)
    return m
}

func (m *Machine) getMem(addr word) byte {
    switch addr {
        case addr < 0x2000:
            return m.mem[addr & 0x7ff]
        case addr < 0x4000:
            //ppu
            break
        case addr < 0x4018:
            //apu etc
            break
        case addr < 0x8000:
            rom.prg_ram[addr-0x6000]
        default:
            return rom.prg_rom
    }
}

func (m *Machine) setMem(addr word, val byte) {
}

func (m *Machine) Run() {
}

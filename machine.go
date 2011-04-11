package gones

import (
"os"
"fmt"
)

type Machine struct {
    cpu *CPU
    ppu *PPU
    rom *ROM
    mem [0x800]byte
}

func MakeMachine(romname string) *Machine {
    m := &Machine{}
    m.cpu = makeCPU(m)
    m.rom = &ROM{}
    f, _ := os.Open(romname, 0, 0)
    m.rom.loadRom(f)
    return m
}

func (m *Machine) getMem(addr word) byte {
    switch true {
        case addr < 0x2000:
            return m.mem[addr & 0x7ff]
        case addr < 0x4000:
            //ppu
            return 0
        case addr < 0x4018:
            //apu etc
            return 0
        case addr < 0x8000:
            return m.rom.prg_ram[addr-0x6000]
        default:
            return m.rom.prg_rom[addr&0x4000 >> 14][addr&0x3fff]
    }
    return 0 //wtf go?
}

func (m *Machine) setMem(addr word, val byte) {
    switch true {
        case addr < 0x2000:
            m.mem[addr & 0x7ff] = val
        case addr < 0x4000:
            //ppu
        case addr < 0x4018:
            //apu etc
        case addr < 0x8000:
            m.rom.prg_ram[addr-0x6000] = val
        default:
            break //mapper write
    }
}

func (m *Machine) Run() {
    m.cpu.reset()
    var inst = Instruction{}
    pc := word(0)
    for true {
        pc = m.cpu.pc
        inst = m.cpu.nextInstruction()
        fmt.Printf("%X  %v %s\n", pc, inst, m.cpu.regs())
        m.cpu.runInstruction(&inst)
    }
}

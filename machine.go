package gones

import (
"⚛sdl"
"os"
"fmt"
)

var keymap = []int{sdl.K_z,
                    sdl.K_x,
                    sdl.K_s,
                    sdl.K_RETURN,
                    sdl.K_UP,
                    sdl.K_DOWN,
                    sdl.K_LEFT,
                    sdl.K_RIGHT}

type Machine struct {
    cpu *CPU
    ppu *PPU
    rom *ROM
    mem [0x800]byte
    read_input_state byte
    keys [8]byte
}

func MakeMachine(romname string) *Machine {
    m := &Machine{}
    m.rom = &ROM{}
    f, _ := os.Open(romname, 0, 0)
    m.rom.loadRom(f)
    m.cpu = makeCPU(m)
    m.ppu = makePPU(m)
    for i := 0; i < 0x800; i++ {
        m.mem[i] = 0xff
    }
    return m
}

func (m *Machine) getMem(addr word) byte {
    switch true {
        case addr < 0x2000:
            return m.mem[addr & 0x7ff]
        case addr < 0x4000:
            return m.ppu.readRegister(int(addr & 7))
        case addr < 0x4018:
            switch addr {
                case 0x4016:
                    if m.read_input_state < 8 {
                        m.read_input_state++
                        return m.keys[m.read_input_state-1]
                    } else {
                        return 1
                    }
            }
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
            m.ppu.writeRegister(int(addr & 7), val)
        case addr < 0x4018:
            switch addr {
                case 0x4016:
                    if val & 1 != 0 {
                        keys := []uint8(sdl.GetKeyState())
                        for i := 0; i < 8; i++ {
                            m.keys[i] = keys[keymap[i]]
                        }
                        m.read_input_state = 0
                    }
                case 0x4014:
                    for v := word(0); v < 0x100; v++ {
                        addr := (v + m.ppu.objAddr) & 0xff
                        m.ppu.objMem[addr] = m.mem[(word(val) << 8) | v]
                    }
            }
            //apu etc
        case addr < 0x8000:
            m.rom.prg_ram[addr-0x6000] = val
        default:
            m.rom.mapper.prgWrite(addr, val)
    }
}

func (m *Machine) Run() {
    m.cpu.reset()
    var inst = Instruction{}
    pc := word(0)
    debug := false
    for true {
        pc = m.cpu.pc
        inst = m.cpu.nextInstruction()
        if debug {
            fmt.Printf("%X  %v %s %s\n", pc, inst, m.cpu.regs(), m.ppu.dump())
        }
        m.cpu.runInstruction(&inst)
        m.ppu.setNTMirroring(m.rom.mirror)
        m.ppu.run()
    }
}

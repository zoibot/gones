package gones

import (
    "⚛sdl"
    "os"
    "fmt"
)

type Machine struct {
    cpu              *CPU
    ppu              *PPU
    apu              *APU
    rom              *ROM
    mem              [0x800]byte
    input            chan []byte
    read_input_state byte
    keys             []byte
    //interrupts
    scheduledNMI     int
}

func MakeMachine(romname string, frames chan []int, input chan []byte) *Machine {
    m := &Machine{input: input}
    m.rom = &ROM{}
    f, err := os.OpenFile(romname, 0, 0)
    if f == nil {
        fmt.Printf("Couldn't open rom!\n%v\n", err)
    }
    m.rom.loadRom(f)
    m.cpu = makeCPU(m)
    m.ppu = makePPU(m, frames)
    m.apu = makeAPU(m)
    m.keys = make([]byte, 8)
    for i := 0; i < 0x800; i++ {
        m.mem[i] = 0xff
    }
    return m
}

func (m *Machine) getMem(addr word) byte {
    switch true {
    case addr < 0x2000:
        return m.mem[addr&0x7ff]
    case addr < 0x4000:
        m.ppu.run()
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
        default:
            return m.apu.readRegister(byte(addr - 0x4000))
        }
        //apu etc
        return 0
    case addr < 0x6000:
        return 0
    case addr < 0x8000:
        return m.rom.prg_ram[addr-0x6000]
    default:
        //TODO the literal 14 here doesn't work anymore!!!
        return m.rom.prg_rom[addr&m.rom.prg_bank_size>>14][addr&(m.rom.prg_bank_size-1)]
    }
    return 0 //wtf go?
}

func (m *Machine) setMem(addr word, val byte) {
    switch true {
    case addr < 0x2000:
        m.mem[addr&0x7ff] = val
    case addr < 0x4000:
        m.ppu.run()
        m.ppu.writeRegister(int(addr&7), val)
    case addr < 0x4018:
        switch addr {
        case 0x4016:
            if val&1 != 0 {
                m.keys = <-m.input
                m.read_input_state = 0
            }
        case 0x4014:
            for v := word(0); v < 0x100; v++ {
                addr := (v + m.ppu.objAddr) & 0xff
                m.ppu.objMem[addr] = m.mem[(word(val)<<8)|v]
            }
        default:
            m.apu.writeRegister(byte(addr - 0x4000), val)
        }
        //apu etc
    case addr < 0x6000:
        return
    case addr < 0x8000:
        m.rom.prg_ram[addr-0x6000] = val
    default:
        m.rom.mapper.prgWrite(addr, val)
    }
}

func (m *Machine) requestNMI() {
    m.scheduledNMI = 2
}

func (m *Machine) suppressNMI() {
    m.scheduledNMI = -1
}

func (m *Machine) requestIrq() {
}

func (m *Machine) runInterrupts() {
    if m.scheduledNMI >= 0 {
        m.scheduledNMI -= 1
    }
    if m.scheduledNMI == 0 {
        m.cpu.nmi()
    }
}

func (m *Machine) syncPPU(cycles uint) {
    m.cpu.cycleCount += uint64(cycles)
    m.ppu.run()
}

func (m *Machine) Debug(keysym uint32) {
    switch keysym {
    case sdl.K_d:
        m.ppu.dumpNTs()
    }
}

func (m *Machine) Run(debug bool) {
    m.cpu.reset()
    var inst = Instruction{}
    pc := word(0)
    for true {
        pc = m.cpu.pc
        inst = m.cpu.nextInstruction()
        if pc == 0xe345 {
        }
        if debug {
            fmt.Printf("%X  %v %s %s\n", pc, inst, m.cpu.regs(), m.ppu.dump())
        }
        cycles := m.cpu.runInstruction(&inst)
        m.ppu.setNTMirroring(m.rom.mirror)
        m.ppu.run()
        m.apu.update(cycles)
        m.runInterrupts()

        //special handling for blargg tests
        if(m.rom.prg_ram[1] == 0xde && m.rom.prg_ram[2] == 0xb0) {
            switch(m.rom.prg_ram[0]) {
                case 0x80:
                    //test running
                case 0x81:
                    //need reset
                default:
                    fmt.Println("test done")
                    i := 0
                    for i = 4; i < len(m.rom.prg_ram)-4; i++ {
                        if m.rom.prg_ram[i] == 0x0 {
                            break
                        }
                    }
                    fmt.Println(string(m.rom.prg_ram[4:i]))
                    os.Exit(0)
            }
        }
    }
}

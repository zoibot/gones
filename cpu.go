package gones

import "fmt"

const (
    N byte = 1 << 7
    V byte = 1 << 6
    B byte = 1 << 4
    D byte = 1 << 3
    I byte = 1 << 2
    Z byte = 1 << 1
    C byte = 1 << 0
)

type CPU struct {
    a, x, y, s, p byte
    pc            word
    scheduledIrq  int
    irqWaiting    bool
    cycleCount    uint64
    m             *Machine
}

func makeCPU(m *Machine) *CPU {
    return &CPU{0,0,0,0,0x24,0,0,false,0,m}
}

func (c *CPU) regs() string {
    return fmt.Sprintf("A:%2X X:%2X Y:%2X P:%2X SP:%2X ", c.a, c.x, c.y, c.p, c.s)
}

func (c *CPU) reset() {
    c.s -= 3
    c.p |= 0x04
    c.pc = wordFromBytes(c.m.getMem(0xfffd), c.m.getMem(0xfffc))
    //apu stuff
}

func (c *CPU) setFlag(flag byte, val bool) {
    if val {
        c.p |= flag
    } else {
        c.p &= ^flag
    }
}

func (c *CPU) getFlag(flag byte) bool {
    return flag & c.p != 0
}

func (c *CPU) setNZ(val byte) {
    c.setFlag(Z, val == 0)
    c.setFlag(N, val & 0x80 != 0)
}

func (c *CPU) push2(val word) {
    c.s -= 2
    var ss word = 0x100
    c.m.setMem(ss | word(c.s+1), byte(val & 0xff))
    c.m.setMem(ss | word(c.s+2), byte(val >> 8))
}

func (c *CPU) pop2() word {
    c.s += 2
    var r word = wordFromBytes(c.m.getMem(word(c.s) | 0x100), c.m.getMem(word((c.s-1) & 0xff) | 0x100))
    return r
}

func (c *CPU) push(val byte) {
    c.s -= 1
    c.m.setMem(word(c.s+1) | 0x100, val)
}

func (c *CPU) pop() byte {
    c.s += 1
    return c.m.getMem(word(c.s) | 0x100)
}

func (c *CPU) nextByte() byte {
    c.pc++
    return c.m.getMem(c.pc - 1)
}

func (c *CPU) nextWordArgs() (word, byte, byte) {
    lo := c.m.getMem(c.pc)
    c.pc++
    hi := c.m.getMem(c.pc)
    c.pc++
    return wordFromBytes(hi,lo), hi, lo
}

func (c *CPU) irq() {
    c.push2(c.pc)
    c.push(c.p)
    c.setFlag(I, true)
    c.pc = wordFromBytes(c.m.getMem(0xffff), c.m.getMem(0xfffe))
    c.cycleCount += 7
}

func (c *CPU) nmi() {
    c.push2(c.pc)
    c.push(c.p)
    c.setFlag(I, true)
    c.pc = wordFromBytes(c.m.getMem(0xfffb), c.m.getMem(0xfffa))
    c.cycleCount += 7
}

func (c *CPU) branch(cond bool, inst *Instruction) {
    if cond {
        inst.extra_cycles += 1
        if inst.addr & 0xff00 != c.pc & 0xff00 {
            inst.extra_cycles += 1
        }
        c.pc = inst.addr
    }
}

func (c *CPU) compare(a byte, b byte) {
    var sa int8 = int8(a)
    var sb int8 = int8(b)
    c.setFlag(N, sa-sb < 0)
    c.setFlag(Z, sa == sb)
    c.setFlag(C, a >= b)
}

func (c *CPU) runInstruction(inst *Instruction) int {
    var (
    m byte = 0
    a7 byte = 0
    m7 byte = 0
    r7 byte = 0
    result word = 0
    )
    switch inst.op.op {
    case NOP, DOP, TOP:
    case JMP:
        c.pc = inst.addr
    case JSR:
        c.push2(c.pc-1)
        c.pc = inst.addr
    case RTS:
        c.pc = c.pop2()+1
    case RTI:
        c.p = (c.pop() | (1<<5)) & (^B)
        c.pc = c.pop2()
		if c.getFlag(I) {
			c.scheduledIrq = 0
		} else if c.irqWaiting {
			c.scheduledIrq = 1
        }
    case BRK:
        c.pc += 1
        c.p |= B
        c.irq()
    case BCS:
        c.branch(c.getFlag(C), inst)
    case BCC:
        c.branch(!c.getFlag(C), inst)
    case BEQ:
        c.branch(c.getFlag(Z), inst)
    case BNE:
        c.branch(!c.getFlag(Z), inst)
    case BVS:
        c.branch(c.getFlag(V), inst)
    case BVC:
        c.branch(!c.getFlag(V), inst)
    case BPL:
        c.branch(!c.getFlag(N), inst)
    case BMI:
        c.branch(c.getFlag(N), inst)
    case BIT:
        m = inst.operand
        c.setFlag(N, m & (1 << 7) != 0)
        c.setFlag(V, m & (1 << 6) != 0)
        c.setFlag(Z, (m & c.a) == 0)
    case CMP:
        c.compare(c.a, inst.operand)
    case CPY:
        c.compare(c.y, inst.operand)
    case CPX:
        c.compare(c.x, inst.operand)
    case CLC:
        c.setFlag(C, false)
    case CLD:
        c.setFlag(D, false)
    case CLV:
        c.setFlag(V, false)
    case CLI:
        c.setFlag(I, false)
    case SED:
        c.setFlag(D, true)
    case SEC:
        c.setFlag(C, true)
    case SEI:
        c.setFlag(I, true)
    case LDA:
        c.a = inst.operand
        c.setNZ(c.a)
    case STA:
        c.m.setMem(inst.addr, c.a)
    case LDX:
        c.x = inst.operand
        c.setNZ(c.x)
    case STX:
        c.m.setMem(inst.addr, c.x)
    case LDY:
        c.y = inst.operand
        c.setNZ(c.y)
    case STY:
        c.m.setMem(inst.addr, c.y)
    case LAX:
        c.a = inst.operand
        c.x = inst.operand
        c.setNZ(c.a)
    case SAX:
        c.m.setMem(inst.addr, c.a & c.x)
    case PHP:
        c.push(c.p | B)
    case PLP:
        c.p = (c.pop() | (1 << 5)) & (^B)
    case PLA:
        c.a = c.pop()
        c.setNZ(c.a)
    case PHA:
        c.push(c.a)
    case AND:
        c.a &= inst.operand
        c.setNZ(c.a)
	case AAC:
		c.a &= inst.operand
		c.setNZ(c.a)
		c.setFlag(C, c.a & 0x80 != 0)
    case ORA:
        c.a |= inst.operand
        c.setNZ(c.a)
    case EOR:
        c.a ^= inst.operand
        c.setNZ(c.a)
    case ADC:
        a7 = c.a & (1 << 7)
        m7 = inst.operand & (1 << 7)
        result = word(c.a) + word(inst.operand)
        if(c.getFlag(C)) {
            result += 1
        }
        c.a = byte(result & 0xff)
        c.setFlag(C, result > 0xff)
        c.setNZ(c.a)
        r7 = c.a & (1 << 7)
        c.setFlag(V, !((a7 != m7) || ((a7 == m7) && (m7 == r7))))
    case SBC:
        a7 = c.a & (1 << 7)
        m7 = inst.operand & (1 << 7)
        result = word(c.a) - word(inst.operand)
        if(!c.getFlag(C)) {
            result -= 1
        }
        c.a = byte(result & 0xff)
        c.setFlag(C, result < 0x100)
        c.setNZ(c.a)
        r7 = c.a & (1 << 7)
        c.setFlag(V, !((a7 == m7) || ((a7 != m7) && (r7 == a7))))
    case INX:
        c.x += 1
        c.setNZ(c.x)
    case INY:
        c.y += 1
        c.setNZ(c.y)
    case DEX:
        c.x -= 1
        c.setNZ(c.x)
    case DEY:
        c.y -= 1
        c.setNZ(c.y)
    case INC:
        inst.operand += 1
        inst.operand &= 0xff
        c.setNZ(inst.operand)
        c.m.setMem(inst.addr, inst.operand)
    case DEC:
        inst.operand -= 1
        inst.operand &= 0xff
        c.setNZ(inst.operand)
        c.m.setMem(inst.addr, inst.operand)
    case DCP:
        c.m.setMem(inst.addr, (inst.operand -1) & 0xff)
        c.compare(c.a, (inst.operand-1)&0xff)
    case ISB:
        inst.operand = (inst.operand + 1) & 0xff
        c.m.setMem(inst.addr, inst.operand)
		a7 = c.a & (1 << 7)
        m7 = inst.operand & (1 << 7)
        result = word(c.a) - word(inst.operand)
        if(!c.getFlag(C)) {
            result -= 1
        }
        c.a = byte(result & 0xff)
        c.setFlag(C, result < 0x100)
        c.setNZ(c.a)
        r7 = c.a & (1 << 7)
        c.setFlag(V, !((a7 == m7) || ((a7 != m7) && (r7 == a7))))
    case LSR_A:
        c.setFlag(C, c.a & 1 != 0)
        c.a >>= 1
        c.setNZ(c.a)
    case LSR:
        c.setFlag(C, inst.operand & 1 != 0)
        inst.operand >>= 1
        c.m.setMem(inst.addr, inst.operand)
        c.setNZ(inst.operand)
    case ASL_A:
        c.setFlag(C, c.a & (1 << 7) != 0)
        c.a <<= 1
        c.setNZ(c.a)
    case ASL:
        c.setFlag(C, inst.operand & (1 << 7) != 0)
        inst.operand <<= 1
        c.m.setMem(inst.addr, inst.operand)
        c.setNZ(inst.operand)
	case ASR:
		c.a &= inst.operand
		c.setFlag(C, c.a & 1 != 0)
		c.a >>= 1
		c.setNZ(c.a)
	case ARR:
		c.a &= inst.operand
		c.a >>= 1
		if c.getFlag(C) {
			c.a |= 0x80
        }
		c.setFlag(C, c.a & (1<<6) != 0)
		c.setFlag(V, ((c.a & (1<<5))<<1)  ^ (c.a & (1<<6)) != 0)
		c.setNZ(c.a)
    case TSX:
        c.x = c.s
        c.setNZ(c.x)
    case TXS:
        c.s = c.x
    case TYA:
        c.a = c.y
        c.setNZ(c.a)
    case TXA:
        c.a = c.x
        c.setNZ(c.a)
	case ATX:
		c.a |= 0xff
		c.a &= inst.operand
		c.x = c.a
		c.setNZ(c.x)
	case AXS:
		c.x = c.a & c.x
		result = word(c.x) - word(inst.operand)
		c.setFlag(C, result < 0x100)
		c.x = byte(result & 0xff)
		c.setNZ(c.x)
	case SYA:
		m = c.y & (inst.args[1] + 1)
		if inst.extra_cycles == 0 {
			c.m.setMem(inst.addr, m)
        }
	case SXA:
		m = c.x & (inst.args[1] + 1)
		if inst.extra_cycles == 0 {
			c.m.setMem(inst.addr, m)
        }
    case ROR_A:
        m = c.a & 1
        c.a >>= 1
        if c.getFlag(C) {
            c.a |= 1 << 7
        }
        c.setFlag(C, m != 0)
        c.setNZ(c.a)
    case ROR:
        m = inst.operand & 1
        inst.operand >>= 1
        if c.getFlag(C) {
            inst.operand |= 1 << 7
        }
        c.setFlag(C, m != 0)
        c.m.setMem(inst.addr, inst.operand)
        c.setNZ(inst.operand)
    case ROL_A:
        m = c.a & (1 << 7)
        c.a <<= 1
        if c.getFlag(C) {
            c.a |= 1
        }
        c.setFlag(C, m != 0)
        c.setNZ(c.a)
    case ROL:
        m = inst.operand & (1 << 7)
        inst.operand <<= 1
        if c.getFlag(C) {
            inst.operand |= 1
        }
        c.setFlag(C, m != 0)
        c.m.setMem(inst.addr, inst.operand)
        c.setNZ(inst.operand)
    case TAY:
        c.y = c.a
        c.setNZ(c.y)
    case TAX:
        c.x = c.a
        c.setNZ(c.x)
    case RLA:
        m = inst.operand & (1 << 7)
        inst.operand <<= 1
        if c.getFlag(C) {
            inst.operand |= 1
        }
        c.setFlag(C, m != 0)
        c.m.setMem(inst.addr, inst.operand)
        c.a &= inst.operand
        c.setNZ(c.a)
    case SLO:
        c.setFlag(C, inst.operand & (1 << 7) != 0)
        inst.operand <<= 1
        c.m.setMem(inst.addr, inst.operand)
        c.a |= inst.operand
        c.setNZ(c.a)
	case SRE:
		c.setFlag(C, inst.operand & 1 != 0)
        inst.operand >>= 1
        c.m.setMem(inst.addr, inst.operand)
        c.a ^= inst.operand
        c.setNZ(c.a)
    case RRA:
        m = inst.operand & 1
        inst.operand >>= 1
        if c.getFlag(C) {
            inst.operand |= 1 << 7
        }
        c.m.setMem(inst.addr, inst.operand)
		a7 = c.a & (1 << 7)
        m7 = inst.operand & (1 << 7)
        result = word(c.a) + word(inst.operand)
        if m != 0  {
            result += 1
        }
        c.a = byte(result & 0xff)
        c.setFlag(C, result > 0xff)
        c.setNZ(c.a)
        r7 = c.a & (1 << 7)
        c.setFlag(V, !((a7 != m7) || ((a7 == m7) && (m7 == r7))))
    case XAA:
    case AXA:
    case XAS:
    case LAR:
    default:
        fmt.Printf("Unsupported opcode! %d", int(inst.op.op))
    }

    c.cycleCount += uint64(inst.op.cycles + inst.extra_cycles)
    return inst.op.cycles + inst.extra_cycles
}

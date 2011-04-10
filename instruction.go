package gones

type op int
const (
BRK op = iota;
ORA; ASL; SLO; PHP; ASL_A; NOP; BPL; CLC; JSR; AND; RLA; TSX;
BIT; ROL; PLP; ROL_A; BMI; SEC; RTI; EOR; SRE; PHA; LSR; PLA;
LSR_A; JMP; BVC; CLI; RTS; ADC; ROR_A; BVS; RRA; SEI; ROR;
STA; SAX; STY; STX; DEY; DEX; TXA; BCC; TYA; TXS; LDY; LDA;
LDX; LAX; TAY; TAX; BCS; CLV; CPY; CPX; CMP; DEC; DCP; INY;
INX; BNE; CLD; SBC; ISB; INC; BEQ; SED; DOP; AAC; ASR; ARR;
ATX; AXS; TOP; SYA; KIL; XAA; AXA; XAS; SXA; LAR;
)

type address_mode int
const (
ZP address_mode = iota
ZP_ST
ZPX
ZPY
IMM
IMP
ABS
REL
A
IXID
IDIX
ABSI
ABSY
ABSX
ABS_ST
)

type Opcode struct {
    op op
    addr_mode address_mode
    illegal bool
    cycles int
    store bool
    extra_page_cross int
}

func makeOpcodeF(o op, a address_mode, c int, epc int) Opcode {
    st := o == STA || o == STX || o == STY
    return Opcode{o,a,false,c,st,epc}
}

func makeOpcodeS(o op, a address_mode, c int) Opcode {
    return makeOpcodeF(o,a,c,1)
}

var Opcodes = [0x100]Opcode{
makeOpcodeS(BRK, IMP, 7),
makeOpcodeS(ORA, IXID, 6),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeS(SLO, IXID, 8),
makeOpcodeS(NOP, ZP, 3),
makeOpcodeS(ORA, ZP, 3),
makeOpcodeS(ASL, ZP, 5),
makeOpcodeS(SLO, ZP, 5),
makeOpcodeS(PHP, IMP, 3),
makeOpcodeS(ORA, IMM, 2),
makeOpcodeS(ASL_A, A, 2),
makeOpcodeS(AAC, IMM, 2),
makeOpcodeS(TOP, ABS, 4),
makeOpcodeS(ORA, ABS, 4),
makeOpcodeS(ASL, ABS, 6),
makeOpcodeS(SLO, ABS, 6),
makeOpcodeS(BPL, REL, 2),
makeOpcodeS(ORA, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(SLO, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(ORA, ZPX, 4),
makeOpcodeS(ASL, ZPX, 6),
makeOpcodeS(SLO, ZPX, 6),
makeOpcodeS(CLC, IMP, 2),
makeOpcodeS(ORA, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(SLO, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(ORA, ABSX, 4),
makeOpcodeF(ASL, ABSX, 7, 0),
makeOpcodeF(SLO, ABSX, 7, 0),
makeOpcodeS(JSR, ABS, 6),
makeOpcodeS(AND, IXID, 6),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeS(RLA, IXID, 8),
makeOpcodeS(BIT, ZP, 3),
makeOpcodeS(AND, ZP, 3),
makeOpcodeS(ROL, ZP, 5),
makeOpcodeS(RLA, ZP, 5),
makeOpcodeS(PLP, IMP, 4),
makeOpcodeS(AND, IMM, 2),
makeOpcodeS(ROL_A, A, 2),
makeOpcodeS(AAC, IMM, 2),
makeOpcodeS(BIT, ABS, 4),
makeOpcodeS(AND, ABS, 4),
makeOpcodeS(ROL, ABS, 6),
makeOpcodeS(RLA, ABS, 6),
makeOpcodeS(BMI, REL, 2),
makeOpcodeS(AND, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(RLA, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(AND, ZPX, 4),
makeOpcodeS(ROL, ZPX, 6),
makeOpcodeS(RLA, ZPX, 6),
makeOpcodeS(SEC, IMP, 2),
makeOpcodeS(AND, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(RLA, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(AND, ABSX, 4),
makeOpcodeF(ROL, ABSX, 7, 0),
makeOpcodeF(RLA, ABSX, 7, 0),
makeOpcodeS(RTI, IMP, 6),
makeOpcodeS(EOR, IXID, 6),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeS(SRE, IXID, 8),
makeOpcodeS(DOP, ZP, 3),
makeOpcodeS(EOR, ZP, 3),
makeOpcodeS(LSR, ZP, 5),
makeOpcodeS(SRE, ZP, 5),
makeOpcodeS(PHA, IMP, 3),
makeOpcodeS(EOR, IMM, 2),
makeOpcodeS(LSR_A, A, 2),
makeOpcodeS(ASR, IMM, 2),
makeOpcodeS(JMP, ABS, 3),
makeOpcodeS(EOR, ABS, 4),
makeOpcodeS(LSR, ABS, 6),
makeOpcodeS(SRE, ABS, 6),
makeOpcodeS(BVC, REL, 2),
makeOpcodeS(EOR, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(SRE, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(EOR, ZPX, 4),
makeOpcodeS(LSR, ZPX, 6),
makeOpcodeS(SRE, ZPX, 6),
makeOpcodeS(CLI, IMP, 2),
makeOpcodeS(EOR, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(SRE, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(EOR, ABSX, 4),
makeOpcodeF(LSR, ABSX, 7, 0),
makeOpcodeF(SRE, ABSX, 7, 0),
makeOpcodeS(RTS, IMP, 6),
makeOpcodeS(ADC, IXID, 6),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeS(RRA, IXID, 8),
makeOpcodeS(DOP, ZP, 3),
makeOpcodeS(ADC, ZP, 3),
makeOpcodeS(ROR, ZP, 5),
makeOpcodeS(RRA, ZP, 5),
makeOpcodeS(PLA, IMP, 4),
makeOpcodeS(ADC, IMM, 2),
makeOpcodeS(ROR_A, A, 2),
makeOpcodeS(ARR, IMM, 2),
makeOpcodeS(JMP, ABSI, 5),
makeOpcodeS(ADC, ABS, 4),
makeOpcodeS(ROR, ABS, 6),
makeOpcodeS(RRA, ABS, 6),
makeOpcodeS(BVS, REL, 2),
makeOpcodeS(ADC, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(RRA, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(ADC, ZPX, 4),
makeOpcodeS(ROR, ZPX, 6),
makeOpcodeS(RRA, ZPX, 6),
makeOpcodeS(SEI, IMP, 2),
makeOpcodeS(ADC, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(RRA, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(ADC, ABSX, 4),
makeOpcodeF(ROR, ABSX, 7, 0),
makeOpcodeF(RRA, ABSX, 7, 0),
makeOpcodeS(DOP, IMM, 2),
makeOpcodeF(STA, IXID, 6, 0),
makeOpcodeS(DOP, IMM, 2),
makeOpcodeS(SAX, IXID, 6),
makeOpcodeS(STY, ZP_ST, 3),
makeOpcodeS(STA, ZP_ST, 3),
makeOpcodeS(STX, ZP_ST, 3),
makeOpcodeS(SAX, ZP, 3),
makeOpcodeS(DEY, IMP, 2),
makeOpcodeS(DOP, IMM, 2),
makeOpcodeS(TXA, IMP, 2),
makeOpcodeS(XAA, IMM, 2),
makeOpcodeS(STY, ABS_ST, 4),
makeOpcodeS(STA, ABS_ST, 4),
makeOpcodeS(STX, ABS_ST, 4),
makeOpcodeS(SAX, ABS, 4),
makeOpcodeS(BCC, REL, 2),
makeOpcodeF(STA, IDIX, 6, 0),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(AXA, IDIX, 6, 0),
makeOpcodeS(STY, ZPX, 4),
makeOpcodeS(STA, ZPX, 4),
makeOpcodeS(STX, ZPY, 4),
makeOpcodeS(SAX, ZPY, 4),
makeOpcodeS(TYA, IMP, 2),
makeOpcodeF(STA, ABSY, 5, 0),
makeOpcodeS(TXS, IMP, 2),
makeOpcodeF(XAS, ABSY, 5, 0),
makeOpcodeF(SYA, ABSX, 5, 0),
makeOpcodeF(STA, ABSX, 5, 0),
makeOpcodeF(SXA, ABSY, 5, 0),
makeOpcodeF(AXA, ABSY, 5, 0),
makeOpcodeS(LDY, IMM, 2),
makeOpcodeS(LDA, IXID, 6),
makeOpcodeS(LDX, IMM, 2),
makeOpcodeS(LAX, IXID, 6),
makeOpcodeS(LDY, ZP, 3),
makeOpcodeS(LDA, ZP, 3),
makeOpcodeS(LDX, ZP, 3),
makeOpcodeS(LAX, ZP, 3),
makeOpcodeS(TAY, IMP, 2),
makeOpcodeS(LDA, IMM, 2),
makeOpcodeS(TAX, IMP, 2),
makeOpcodeS(ATX, IMM, 2),
makeOpcodeS(LDY, ABS, 4),
makeOpcodeS(LDA, ABS, 4),
makeOpcodeS(LDX, ABS, 4),
makeOpcodeS(LAX, ABS, 4),
makeOpcodeS(BCS, REL, 2),
makeOpcodeS(LDA, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeS(LAX, IDIX, 5),
makeOpcodeS(LDY, ZPX, 4),
makeOpcodeS(LDA, ZPX, 4),
makeOpcodeS(LDX, ZPY, 4),
makeOpcodeS(LAX, ZPY, 4),
makeOpcodeS(CLV, IMP, 2),
makeOpcodeS(LDA, ABSY, 4),
makeOpcodeS(TSX, IMP, 2),
makeOpcodeS(LAR, ABSY, 4),
makeOpcodeS(LDY, ABSX, 4),
makeOpcodeS(LDA, ABSX, 4),
makeOpcodeS(LDX, ABSY, 4),
makeOpcodeS(LAX, ABSY, 4),
makeOpcodeS(CPY, IMM, 2),
makeOpcodeS(CMP, IXID, 6),
makeOpcodeS(DOP, IMM, 2),
makeOpcodeS(DCP, IXID, 8),
makeOpcodeS(CPY, ZP, 3),
makeOpcodeS(CMP, ZP, 3),
makeOpcodeS(DEC, ZP, 5),
makeOpcodeS(DCP, ZP, 5),
makeOpcodeS(INY, IMP, 2),
makeOpcodeS(CMP, IMM, 2),
makeOpcodeS(DEX, IMP, 2),
makeOpcodeS(AXS, IMM, 2),
makeOpcodeS(CPY, ABS, 4),
makeOpcodeS(CMP, ABS, 4),
makeOpcodeS(DEC, ABS, 6),
makeOpcodeS(DCP, ABS, 6),
makeOpcodeS(BNE, REL, 2),
makeOpcodeS(CMP, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(DCP, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(CMP, ZPX, 4),
makeOpcodeS(DEC, ZPX, 6),
makeOpcodeS(DCP, ZPX, 6),
makeOpcodeS(CLD, IMP, 2),
makeOpcodeS(CMP, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(DCP, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(CMP, ABSX, 4),
makeOpcodeF(DEC, ABSX, 7, 0),
makeOpcodeF(DCP, ABSX, 7, 0),
makeOpcodeS(CPX, IMM, 2),
makeOpcodeS(SBC, IXID, 6),
makeOpcodeS(DOP, IMM, 2),
makeOpcodeS(ISB, IXID, 8),
makeOpcodeS(CPX, ZP, 3),
makeOpcodeS(SBC, ZP, 3),
makeOpcodeS(INC, ZP, 5),
makeOpcodeS(ISB, ZP, 5),
makeOpcodeS(INX, IMP, 2),
makeOpcodeS(SBC, IMM, 2),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeS(SBC, IMM, 2),
makeOpcodeS(CPX, ABS, 4),
makeOpcodeS(SBC, ABS, 4),
makeOpcodeS(INC, ABS, 6),
makeOpcodeS(ISB, ABS, 6),
makeOpcodeS(BEQ, REL, 2),
makeOpcodeS(SBC, IDIX, 5),
makeOpcodeS(KIL, IMP, 0),
makeOpcodeF(ISB, IDIX, 8, 0),
makeOpcodeS(DOP, ZPX, 4),
makeOpcodeS(SBC, ZPX, 4),
makeOpcodeS(INC, ZPX, 6),
makeOpcodeS(ISB, ZPX, 6),
makeOpcodeS(SED, IMP, 2),
makeOpcodeS(SBC, ABSY, 4),
makeOpcodeS(NOP, IMP, 2),
makeOpcodeF(ISB, ABSY, 7, 0),
makeOpcodeS(TOP, ABSX, 4),
makeOpcodeS(SBC, ABSX, 4),
makeOpcodeF(INC, ABSX, 7, 0),
makeOpcodeF(ISB, ABSX, 7, 0),
}

type Instruction struct {
    op Opcode
    extra_cycles int
    addr word
    operand byte
    args [2]byte
    arglen int
}

func (o op) String() string {
    return "FOO"
}

func (i *Instruction) String() string {
    return i.op.op.String()
}

func (m *Machine) next_instruction() Instruction {
    opcode := m.nextByte()
    extra_cycles := 0
    op := Opcodes[opcode]
    args := [2]byte{0,0}
    arglen := 0
    var operand byte = 0
    var addr word = 0
    var i_addr word = 0
    switch(op.addr_mode) {
    case IMM:
        operand = m.nextByte()
        args[0] = operand
        arglen = 1
    case ZP:
        addr = word(m.nextByte())
        operand = m.getMem(addr)
        args[0] = byte(addr)
        arglen = 1
    case ZP_ST:
        addr = word(m.nextByte())
        args[0] = byte(addr)
        arglen = 1
    case ABS:
        addr, args[0], args[1] = m.nextWordArgs()
        operand = m.getMem(addr)
        arglen = 2
    case ABS_ST:
        addr, args[0], args[1] = m.nextWordArgs()
        arglen = 2
    case ABSI:
        i_addr, args[0], args[1] = m.nextWordArgs()
        addr = wordFromBytes((m.getMem(i_addr)),
                m.getMem(i_addr+1 & 0xff + i_addr&0xff00))
        arglen = 2
    case ABSY:
        i_addr, args[0], args[1] = m.nextWordArgs()
        addr = i_addr+m.y & 0xffff

    default:
        operand++
    }
    return Instruction{op, extra_cycles, addr, operand, args, arglen}
}

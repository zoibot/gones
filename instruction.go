package instruction

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

func makeOpcode(o op, a address_mode, c int, i bool, epc int) Opcode {
    return Opcode{o,a,i,c,false,epc}
}

var Opcodes = []Opcode{
makeOpcode(BRK, IMP, 0, false, 0),
}

type Instruction struct {
    op Opcode
}

func (o op) String() string {
    return "FOO"
}

func (i *Instruction) String() string {
    return i.op.op.String()
}

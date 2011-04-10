package gones


type Machine struct {
    a, x, y, s, p byte
    pc word
}

func (m *Machine) nextByte() byte {
    m.pc++
    return m.getMem(m.pc-1)
}

func (m *Machine) nextWordArgs() (word, byte, byte) {
    lo := m.getMem(m.pc)
    m.pc++
    hi := m.getMem(m.pc)
    m.pc++
    return word(hi)<<8 + word(lo), hi, lo
}

func (m *Machine) getMem(addr word) byte {
    return 0
}

func (m *Machine) setMem(addr word, val byte) {

}

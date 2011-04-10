package gones

type word uint16

func wordFromBytes(hi byte, lo byte) word {
    return word(hi)<<8 + word(lo)
}

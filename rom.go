package gones

import (
    "os"
    "fmt"
)

type ROM struct {
    fname string
    chr_size byte
    chr_ram bool
    chr_rom [2][]byte
    chr_banks []byte
    prg_ram []byte
    prg_size byte
    prg_rom [2][]byte
    prg_banks []byte
    flags6, flags7 byte
    mapper_num byte
    mapper Mapper
    mirror int
}

func (r *ROM) loadRom(f *os.File) {
    header := make([]byte, 16)
    r.fname = f.Name()
    f.Read(header)
    if string(header[:4]) == "NES\x1a" {
        fmt.Printf("header constant OK!\n")
    } else {
        fmt.Printf("bad rom...\n")
    }
    r.prg_size = header[4]
    r.chr_size = header[5]
    r.flags6 = header[6]
    r.flags7 = header[7]
    if r.flags6 & 1 != 0 {
        r.mirror = VERTICAL
    } else {
        r.mirror = HORIZONTAL
    }
    r.mapper_num = (r.flags7 & 0xf0) | ((r.flags6 & 0xf0)>>4)
    prg_ram_size := header[8]
    if r.flags6 & (1<<2) != 0 {
        fmt.Printf("loading trainer\n")
        trainer := make([]byte, 512)
        f.Read(trainer)
    }
    r.prg_banks = make([]byte, uint(r.prg_size) * 0x4000)
    f.Read(r.prg_banks)
    if r.chr_size != 0 {
        r.chr_banks = make([]byte, uint(r.chr_size) * 0x2000)
        f.Read(r.chr_banks)
    } else {
        r.chr_banks = make([]byte, 0x2000)
    }
    r.mapper = loadMapper(r.mapper_num, r)
    fmt.Printf("prg size %d\nchr size %d\n", r.prg_size, r.chr_size)
    if prg_ram_size == 0 {
        r.prg_ram = make([]byte, 0x4000)
    } else {
        r.prg_ram = make([]byte, uint(prg_ram_size) * 0x4000)
    }
    fmt.Printf("Rom loaded successfully!\n")
}



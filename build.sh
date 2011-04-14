#!/bin/sh

6g -o gones.6 instruction.go machine.go cpu.go util.go ppu.go rom.go mapper.go
6g main.go
6l -o gones main.6

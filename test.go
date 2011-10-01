package main

import (
	"os"
	"fmt"
	"bytes"
	"big"
	"strconv"
	"./gones"
)

var buttonmap = map[byte]byte{
	'A': 0,
	'B': 1,
	'S': 2,
	'T': 3,
	'U': 4,
	'D': 5,
	'L': 6,
	'R': 7}

func testMany(fname string) {
	f, e := os.Open(fname)
	if f == nil {
		fmt.Printf("error opening test file: %v\n", e)
		os.Exit(1)
	}
	finfo, _ := f.Stat()
	sz := finfo.Size
	buf := make([]byte, sz)
	f.Read(buf)
	lines := bytes.Split(buf, []byte{'\n'})
	for namei := range lines {
		test(string(lines[namei]))
	}
}

func test(tfile string) {
	f, e := os.Open(tfile)
	if f == nil {
		fmt.Printf("Error opening test file: %v\n", e)
		os.Exit(1)
	}
	finfo, _ := f.Stat()
	sz := finfo.Size
	buf := make([]byte, sz)
	f.Read(buf)
	lines := bytes.Split(buf, []byte{'\n'})
	romname := lines[0]
	frames := make(chan []int)
	currentInput := make([]byte, 8)
	m := gones.MakeMachine(string(romname), frames,
		func() chan []byte {
			c := make(chan []byte)
			go func() {
				for {
					c <- currentInput
				}
			}()
			return c
		}())
	go m.Run(false)
	for line := range lines[1:] {
		fs := bytes.Fields(lines[line])
		switch string(fs[0]) {
		case "wait":
			length, _ := strconv.Atoi(string(fs[1]))
			//fmt.Printf("waiting %v\n", length)
			for i := 0; i < length; i++ {
				<-frames
			}
		case "screen":
			//fmt.Printf("screenshot\n")
			frame := <-frames
			gones.SaveImage("test/test.png", frame)
		case "test_image":
			frame := <-frames
			hash := big.NewInt(0)
			hash.SetBytes(gones.HashImage(frame))
			if bytes.Compare(fs[1], []byte(fmt.Sprintf("%x", hash))) != 0 {
				fmt.Printf("%s %s: fail %x\n", romname, string(fs[2]), hash)
				gones.SaveImage("test/test.png", frame)
			} else {
				fmt.Printf("%s %s: pass\n", romname, string(fs[2]))
			}
		/*case "blargg_string":
		  str, status, code = gones.GetBlarggOutput()
		  if code != 1 {
		      fmt.Printf("fail")
		  } else {
		      fmt.Printf("pass")
		  }*/
		case "press":
			key := buttonmap[fs[1][0]]
			//fmt.Printf("pressing key %v\n", key)
			currentInput[key] = 1
			<-frames
			<-frames
			currentInput[key] = 0
			//case "test":
		}
	}
}

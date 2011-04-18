package gones

import "image"
import "image/png"
import "crypto/md5"
import "fmt"
import "os"

type word uint16

func wordFromBytes(hi byte, lo byte) word {
    return word(hi)<<8 + word(lo)
}

func min(a int, b int) int {
    if a < b {
        return a
    } else {
        return b
    }
    return 0
}

func intToColor(col int) image.RGBAColor {
    r := byte((col & 0xff0000) >> 16)
    g := byte((col & 0x00ff00) >> 8)
    b := byte((col & 0x0000ff))
    return image.RGBAColor{r, g, b, 0xff}
}

func intsToImage(frame []int) image.Image {
    img := image.NewRGBA(256, 240)
    for y := 0; y < 240; y++ {
        for x := 0; x < 256; x++ {
            img.Set(x, y, intToColor(frame[y*256+x]))
        }
    }
    return img
}

func HashImage(frame []int) []byte {
    img := intsToImage(frame)
    h := md5.New()
    png.Encode(h, img)
    return h.Sum()
}

func SaveImage(fname string, frame []int) {
    img := intsToImage(frame)
    f, err := os.Open(fname, os.O_CREAT|os.O_WRONLY, 0666)
    if f == nil {
        fmt.Printf("error opening file. %v\n", err.String())
    }
    png.Encode(f, img)
}

var colors = []int{
    0x7C7C7C,
    0x0000FC,
    0x0000BC,
    0x4428BC,
    0x940084,
    0xA80020,
    0xA81000,
    0x881400,
    0x503000,
    0x007800,
    0x006800,
    0x005800,
    0x004058,
    0x000000,
    0x000000,
    0x000000,
    0xBCBCBC,
    0x0078F8,
    0x0058F8,
    0x6844FC,
    0xD800CC,
    0xE40058,
    0xF83800,
    0xE45C10,
    0xAC7C00,
    0x00B800,
    0x00A800,
    0x00A844,
    0x008888,
    0x000000,
    0x000000,
    0x000000,
    0xF8F8F8,
    0x3CBCFC,
    0x6888FC,
    0x9878F8,
    0xF878F8,
    0xF85898,
    0xF87858,
    0xFCA044,
    0xF8B800,
    0xB8F818,
    0x58D854,
    0x58F898,
    0x00E8D8,
    0x787878,
    0x000000,
    0x000000,
    0xFCFCFC,
    0xA4E4FC,
    0xB8B8F8,
    0xD8B8F8,
    0xF8B8F8,
    0xF8A4C0,
    0xF0D0B0,
    0xFCE0A8,
    0xF8D878,
    0xD8F878,
    0xB8F8B8,
    0xB8F8D8,
    0x00FCFC,
    0xF8D8F8,
    0x000000,
    0x000000}

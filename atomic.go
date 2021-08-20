package main

// #cgo LDFLAGS: -L. -lffthandler -lm
// #include "FFTHandler.h"
import "C"
import (
    "math"
    "unsafe"
    "sort"
)

type FFTHandler C.FFTHandler

type Coordinate struct {
    X int
    Y int
}

func min(a int, b int) int {
    if a < b { return a }
    return b
}

func squaredDistance(p Coordinate, q Coordinate) int {
    return ((p.X - q.X) * (p.X - q.X)) + ((p.Y - q.Y) * (p.Y - q.Y))
}

func NewFFTHandler(n int32, samplingRate int32) *FFTHandler {
    handler := C.newHandler(C.int(n), C.int(samplingRate))
    return (*FFTHandler)(handler)
}

func (ffth *FFTHandler) Destroy() {
    C.destroyHandler((*C.FFTHandler)(ffth))
}

func (ffth *FFTHandler) Spectrogram(data []float64, stride int32) []float64 {
	buffers := C.spectrogram((*C.double)(unsafe.Pointer(&data[0])), C.int(len(data)), C.int(stride), (*C.FFTHandler)(ffth))
    timesteps := (len(data) - int(ffth.n)) / int(stride)
    length := timesteps * int(ffth.fftCount)
	sliceStruct := struct {
		address uintptr
		length int
		capacity int
	}{uintptr(unsafe.Pointer(buffers)), length, length}
    return *(*[]float64)(unsafe.Pointer(&sliceStruct))
}

func Constellation(spec []float64, n int, start int, cols int) []Coordinate {
    stars := []Coordinate{}
    windows := len(spec) / n
    for x := start; x < start + cols; x++ {
        for y := 0; y < n; y++ {
            max := spec[y + x * n]
            if max < 10 {
                continue
            }
            for yOffset := 0; yOffset < 41; yOffset++ {
                newY := y + (20 - yOffset)
                if newY < 0 || newY >= n {
                    continue
                }
                skip := int(math.Abs(20 - float64(yOffset)))
                for xOffset := skip; xOffset < (41 - skip); xOffset++ {
                    newX := x + (20 - xOffset)
                    if newX < 0 || newX >= windows {
                        continue
                    }
                    if max < spec[newY + newX * n] {
                        max = spec[newY + newX * n]
                    }
                }
            }
            if spec[y + x * n] == max {
                stars = append(stars, Coordinate{x, y})
            }
        }
    }
    return stars
}

func ConstellationParallel(gos int, spec []float64, n int) []Coordinate {
    outChannel := make(chan []Coordinate, gos)
    windows := len(spec) / n
    split := windows / gos
    extraLast := windows % gos
    stars := []Coordinate{}
    for i := 0; i < gos; i++ {
        process := split
        if i == gos - 1 {
            process += extraLast
        }
        go func(output chan []Coordinate, gram []float64, colLen int, splitNum int, cols int, index int) {
            output <- Constellation(gram, colLen, index * splitNum, cols)
        }(outChannel, spec, n, split, process, i)
    }
    for i := 0; i < gos; i++ {
        out := <-outChannel
        stars = append(stars, out...)
    }
    return stars
}

func Fingerprints(stars []Coordinate) map[uint][]int {
    // unint because of bit shifts, more simple
    hashes := make(map[uint][]int)
    // uint is the hashes and []int is the list of timestamps (star.X) where those hashes start at
    for i, star := range stars {
        nextStars := stars[i + 1:min(i + 1001, len(stars))]
        sort.Slice(nextStars, func(j int, k int) bool {
            return squaredDistance(star, nextStars[j]) < squaredDistance(star, nextStars[k])
        })
        closestStars := nextStars[:min(20, len(nextStars))]
        for _, neighbour := range closestStars {
            delta := neighbour.X - star.X
            hash := uint(star.Y) | uint(neighbour.Y << 16) | uint(delta << 32)
            // bits 0 to 15 represents star.Y, bits 16 to 31 represents neighbour.Y, bits 32 to 47 represents delta
            // query hashes map for the slice behind the hash
            // if query is okay, exists, append to slice and insert new slice in map for hash (append star.X)
            // if query is not okay, set hash in map to slice of just [star.X]
            slice, ok := hashes[hash]
            if ok {
                hashes[hash] = append(slice, star.X)
            } else {
                hashes[hash] = []int{star.X}
            }
        }
    }
    return hashes
}

func AudioFingerprints(audio []float64, fftHandler *FFTHandler, stride int32, goroutines int) map[uint][]int {
    s := fftHandler.Spectrogram(audio, stride)
    c := ConstellationParallel(goroutines, s, int(fftHandler.fftCount))
    f := Fingerprints(c)
    return f
}

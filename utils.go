package main

import (
    "fmt"
    "os"
    "io"

    "github.com/youpy/go-wav"
)

func LoadAudioFromFile(filename string, sampleRate int) ([]float64, error) {
    file, err := os.Open(filename)
    if err != nil {
        return []float64{}, err
    }
    defer file.Close()

    reader := wav.NewReader(file)
    format, err := reader.Format()
    if err != nil {
        return []float64{}, err
    }
    originalSampleRate := int(format.SampleRate)

    values := []float64{}
    for {
        samples, err := reader.ReadSamples()
        if err == io.EOF {
            break
        }
        for _, sample := range samples {
            values = append(values, float64(sample.Values[0]))
        }
    }

    if originalSampleRate == sampleRate {
        return values, nil
    } else {
        return ResampleAudio(values, originalSampleRate, sampleRate), nil
    }
}

func Interpolate(sampleX []float64, refX []float64, values []float64) []float64 {
    sampledValues := make([]float64, len(sampleX))
    refIdx := 0
    for sampleIdx, sampleVal := range sampleX {
        for refIdx < len(refX) && refX[refIdx] < sampleVal {
            refIdx++
        }
        if refIdx >= len(refX) - 2 {
            return []float64{}
        }
        xScale := (refX[refIdx] - sampleVal) / (refX[refIdx + 1] - refX[refIdx])
        yScaled := values[refIdx] + (values[refIdx + 1] - values[refIdx]) * xScale
        sampledValues[sampleIdx] = yScaled
    }
    return sampledValues
}

func Linspace(start float64, end float64, count int, endpoint bool) []float64 {
    var gap float64
    if endpoint {
        gap = (end - start) / float64(count - 1)
    } else {
        gap = (end - start) / float64(count)
    }
    values := make([]float64, count)
    for i := 0; i < count; i++ {
        values[i] = start + gap * float64(i)
    }
    return values
}

func ResampleAudio(audio []float64, oldRate int, newRate int) []float64 {
    scale := float64(newRate) / float64(oldRate)
    newLen := int(float64(len(audio)) * scale)
    sampleX := Linspace(0, 10, newLen, false)
    refX := Linspace(0, 10, len(audio), false)
    return Interpolate(sampleX, refX, audio)
}

func UploadQueryFingerprints(fingerprints map[uint][]int, queryID int, authToken string, authSettings AuthSettings, gos int) {
    inChannel := make(chan struct {
        fingerprint uint
        timestamps []int
    }, gos)
    for i := 0; i < gos; i++ {
        go func() {
            for fingerprint := range inChannel {
                for _, timestamp := range fingerprint.timestamps {
                    params := map[string]interface{}{
                        "QueryID": queryID,
                        "Hash": fingerprint.fingerprint,
                        "Time": timestamp,
                    }
                    err := Db2RunSyncJobWithoutResponse(authToken, authSettings, "NewQueryHash", "1.0", params)
                    if err != nil {
                        fmt.Println(err)
                    }
                }
            }
        }()
    }
    for fingerprint, timestamps := range fingerprints {
        inChannel <- struct {
            fingerprint uint
            timestamps []int
        }{fingerprint, timestamps}
    }
    close(inChannel)
}

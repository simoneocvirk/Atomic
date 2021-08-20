package main

import (
    "fmt"
)

const WINDOW_SIZE = 4096
const SAMPLING_RATE = 48000
const STRIDE int32 = 2048
const GOROUTINES = 100

func main() {
    handler := NewFFTHandler(WINDOW_SIZE, 48000)
    defer handler.Destroy()

    authSettings := AuthSettings{
        RestHostname: "52.117.200.43",
        DBHostname: "db2w-bhfzmxl.us-east.db2w.cloud.ibm.com",
        Database: "BLUDB",
        DBPort: 50000,
        RestPort: 50050,
        SSL: false,
        Password: "zVrhUuN3gz@1dgFuZ1XTPbe@j9wH_",
        Username: "bluadmin",
        ExpiryTime: "24h",
    }

    authToken, err := Db2Authenticate(authSettings)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(authToken)

    audio, err := LoadAudioFromFile("empty.wav", 48000)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(len(audio))

    features := AudioFingerprints(audio[48000*5:48000*7], handler, STRIDE, GOROUTINES)
    fmt.Println(len(features))

    UploadQueryFingerprints(features, 5, authToken, authSettings, GOROUTINES)
}

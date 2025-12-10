package main

import (
    "fmt"
    "time"
)

func say(msg string) {
    for i := 0; i < 3; i++ {
        fmt.Println(msg, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    go say("goroutine")
    say("main")
}

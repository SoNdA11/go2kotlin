package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2

    select {
    case v := <-ch:
        fmt.Println("recebi", v)
    default:
        fmt.Println("nenhum valor")
    }

    timeout := time.After(200 * time.Millisecond)
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch <- 3
    }()

    select {
    case v := <-ch:
        fmt.Println("recebi depois", v)
    case <-timeout:
        fmt.Println("timeout")
    }
}

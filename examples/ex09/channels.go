package main

import "fmt"

func producer(ch chan<- int, n int) {
    for i := 0; i < n; i++ {
        ch <- i
    }
    close(ch)
}

func consumer(ch <-chan int) {
    for v := range ch {
        fmt.Println("consumed", v)
    }
}

func main() {
    ch := make(chan int)
    go producer(ch, 5)
    consumer(ch)
}

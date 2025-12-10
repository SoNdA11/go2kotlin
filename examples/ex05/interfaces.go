package main

import "fmt"

type Shape interface {
    Area() float64
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return 3.14159 * c.Radius * c.Radius
}

type Rectangle struct {
    W, H float64
}

func (r Rectangle) Area() float64 {
    return r.W * r.H
}

func printArea(s Shape) {
    fmt.Println("Area:", s.Area())
}

func main() {
    c := Circle{Radius: 2.5}
    r := Rectangle{W: 3, H: 4}
    printArea(c)
    printArea(r)
}

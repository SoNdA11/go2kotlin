package main

import "fmt"

type Point struct {
    X, Y int
}

func (p Point) Move(dx, dy int) Point {
    p.X += dx
    p.Y += dy
    return p
}

func (p *Point) Scale(factor int) {
    p.X *= factor
    p.Y *= factor
}

func main() {
    p := Point{X: 1, Y: 2}
    p2 := p.Move(3, 4)
    p.Scale(2)
    fmt.Println("p:", p)
    fmt.Println("p2:", p2)
}

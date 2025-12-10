package main

import "fmt"

func main() {
    nums := []int{1, 2, 3, 4}
    nums = append(nums, 5)

    m := map[string]int{
        "um": 1,
        "dois": 2,
    }
    m["tres"] = 3

    for i, v := range nums {
        fmt.Println(i, v)
    }

    for k, v := range m {
        fmt.Println(k, v)
    }
}

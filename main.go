package main

import (
    "go.k6.io/xk6"
    _ "github.com/LiorGandin/k6-rti-extension/rti"
)

func main() {
    xk6.Build(xk6.Args{
        Extensions: []xk6.Extension{
            {
                Name:    "github.com/LiorGandin/k6-rti-extension",
                Package: "github.com/LiorGandin/k6-rti-extension/rti",
            },
        },
    })
}

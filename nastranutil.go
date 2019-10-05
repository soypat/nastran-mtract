package main

type dims struct {
	x    float64
	y    float64
	z    float64
	t    float64
	csys int // dont think ill use it (coordinate system)
}

type node struct {
	number int
	dims
}

type element struct {
	number    int
	Type      string
	nodeIndex []int
	collector int
}
type constraint struct {
	number    int
	Type      string
	master    int
	slaves    []int
	collector int
}

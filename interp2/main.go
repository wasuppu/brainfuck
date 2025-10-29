package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var memory [30000]byte

type stack struct {
	items []int
}

func (s *stack) Push(item int) {
	s.items = append(s.items, item)
}

func (s *stack) Pop() int {
	l := len(s.items)
	v := s.items[l-1]
	s.items = s.items[:l-1]
	return v
}

type Ops int

const (
	Left Ops = iota
	Right
	Add
	Sub
	LBrack
	RBrack
	Output
	Input
)

func parse(prog []byte) []Ops {
	ops := []Ops{}
	for _, b := range prog {
		switch b {
		case '<':
			ops = append(ops, Left)
		case '>':
			ops = append(ops, Right)
		case '+':
			ops = append(ops, Add)
		case '-':
			ops = append(ops, Sub)
		case '.':
			ops = append(ops, Output)
		case ',':
			ops = append(ops, Input)
		case '[':
			ops = append(ops, LBrack)
		case ']':
			ops = append(ops, RBrack)
		}
	}
	return ops
}

func preprocess(ops []Ops) map[int]int {
	s := stack{}
	jumpTable := map[int]int{}

	for pos, cmd := range ops {
		switch cmd {
		case LBrack:
			s.Push(pos)
		case RBrack:
			start := s.Pop()
			jumpTable[start] = pos
			jumpTable[pos] = start
		}
	}
	return jumpTable
}

func execute(r io.Reader, i io.Reader, w io.Writer) {
	prog, err := io.ReadAll(r)
	must(err)
	input := bufio.NewReader(i)

	ops := parse(prog)
	jumptable := preprocess(ops)

	ptr := 0
	pos := 0
	for pos < len(ops) {
		switch ops[pos] {
		case Left:
			ptr--
		case Right:
			ptr++
		case Add:
			memory[ptr]++
		case Sub:
			memory[ptr]--
		case Output:
			fmt.Fprintf(w, "%c", memory[ptr])
		case Input:
			if memory[ptr], err = input.ReadByte(); err != nil {
				if err != io.EOF {
					must(err)
				}
			}
		case LBrack:
			if memory[ptr] == 0 {
				pos = jumptable[pos]
			}
		case RBrack:
			if memory[ptr] != 0 {
				pos = jumptable[pos]
			}
		}
		pos++
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file.bf]\n", os.Args[0])
		return
	}

	file, err := os.Open(os.Args[1])
	must(err)
	execute(file, os.Stdin, os.Stdout)
}

func must(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

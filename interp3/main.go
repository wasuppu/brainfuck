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

type OpKind byte

const (
	Left   OpKind = '<'
	Right  OpKind = '>'
	Add    OpKind = '+'
	Sub    OpKind = '-'
	LBrack OpKind = '['
	RBrack OpKind = ']'
	Output OpKind = '.'
	Input  OpKind = ','
)

type Ops struct {
	Kind OpKind
	Jump int
}

func parse(prog []byte) []Ops {
	ops := []Ops{}
	s := stack{}

	for _, b := range prog {
		switch b {
		case '<', '>', '+', '-', '.', ',':
			ops = append(ops, Ops{Kind: OpKind(b)})
		case '[':
			s.Push(len(ops))
			ops = append(ops, Ops{Kind: LBrack})
		case ']':
			start := s.Pop()
			ops[start].Jump = len(ops)
			ops = append(ops, Ops{Kind: RBrack, Jump: start})
		}
	}
	return ops
}

func execute(r io.Reader, i io.Reader, w io.Writer) {
	prog, err := io.ReadAll(r)
	must(err)
	input := bufio.NewReader(i)

	ops := parse(prog)

	ptr := 0
	pos := 0
	for pos < len(ops) {
		op := ops[pos]
		switch op.Kind {
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
				pos = op.Jump
			}
		case RBrack:
			if memory[ptr] != 0 {
				pos = op.Jump
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

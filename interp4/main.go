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
	Kind   OpKind
	Jump   int
	Oprand int
}

func parse(prog []byte) []Ops {
	ops := []Ops{}
	s := stack{}

	for i := 0; i < len(prog); i++ {
		b := prog[i]
		switch b {
		case '<', '>', '+', '-', '.', ',':
			count := 1
			j := i + 1
			for j < len(prog) && prog[j] == b {
				count += 1
				j += 1
			}
			ops = append(ops, Ops{Kind: OpKind(b), Oprand: count})
			i = j - 1
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
			ptr -= op.Oprand
		case Right:
			ptr += op.Oprand
		case Add:
			memory[ptr] = memory[ptr] + byte(op.Oprand)
		case Sub:
			memory[ptr] = memory[ptr] - byte(op.Oprand)
		case Output:
			for range op.Oprand {
				fmt.Fprintf(w, "%c", memory[ptr])
			}
		case Input:
			for range op.Oprand {
				if memory[ptr], err = input.ReadByte(); err != nil {
					if err != io.EOF {
						must(err)
					}
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

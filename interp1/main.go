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

func preprocess(prog []byte) map[int]int {
	s := stack{}
	jumpTable := map[int]int{}

	for pos, cmd := range prog {
		switch cmd {
		case '[':
			s.Push(pos)
		case ']':
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

	jumptable := preprocess(prog)

	ptr := 0
	pos := 0
	for pos < len(prog) {
		switch prog[pos] {
		case '<':
			ptr--
		case '>':
			ptr++
		case '+':
			memory[ptr]++
		case '-':
			memory[ptr]--
		case '[':
			if memory[ptr] == 0 {
				pos = jumptable[pos]
			}
		case ']':
			if memory[ptr] != 0 {
				pos = jumptable[pos]
			}
		case '.':
			fmt.Fprintf(w, "%c", memory[ptr])
		case ',':
			if memory[ptr], err = input.ReadByte(); err != nil {
				if err != io.EOF {
					must(err)
				}
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

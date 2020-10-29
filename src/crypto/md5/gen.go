// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program generates md5block.go
// Invoke as
//
//	go run gen.go -output md5block.go

package main

import (
	"bytes"
	"flag"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

var filename = flag.String("output", "md5block.go", "output file name")

func main() {
	flag.Parse()

	var buf bytes.Buffer

	t := template.Must(template.New("main").Funcs(funcs).Parse(program))
	if err := t.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(*filename, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

type Data struct {
	a, b, c, d string
	Shift1     []int
	Shift2     []int
	Shift3     []int
	Shift4     []int
	Table1     []uint32
	Table2     []uint32
	Table3     []uint32
	Table4     []uint32
}

var funcs = template.FuncMap{
	"dup":     dup,
	"relabel": relabel,
	"rotate":  rotate,
	"idx":     idx,
	"seq":     seq,
}

func dup(count int, x []int) []int {
	var out []int
	for i := 0; i < count; i++ {
		out = append(out, x...)
	}
	return out
}

func relabel(s string) string {
	return strings.NewReplacer("arg0", data.a, "arg1", data.b, "arg2", data.c, "arg3", data.d).Replace(s)
}

func rotate() string {
	data.a, data.b, data.c, data.d = data.d, data.a, data.b, data.c
	return "" // no output
}

func idx(round, index int) int {
	v := 0
	switch round {
	case 1:
		v = index
	case 2:
		v = (1 + 5*index) & 15
	case 3:
		v = (5 + 3*index) & 15
	case 4:
		v = (7 * index) & 15
	}
	return v
}

func seq(i int) []int {
	s := make([]int, i)
	for i := range s {
		s[i] = i
	}
	return s
}

var data = Data{
	a:      "a",
	b:      "b",
	c:      "c",
	d:      "d",
	Shift1: []int{7, 12, 17, 22},
	Shift2: []int{5, 9, 14, 20},
	Shift3: []int{4, 11, 16, 23},
	Shift4: []int{6, 10, 15, 21},

	// table[i] = int((1<<32) * abs(sin(i+1 radians))).
	Table1: []uint32{
		// round 1
		0xd76aa478,
		0xe8c7b756,
		0x242070db,
		0xc1bdceee,
		0xf57c0faf,
		0x4787c62a,
		0xa8304613,
		0xfd469501,
		0x698098d8,
		0x8b44f7af,
		0xffff5bb1,
		0x895cd7be,
		0x6b901122,
		0xfd987193,
		0xa679438e,
		0x49b40821,
	},
	Table2: []uint32{
		// round 2
		0xf61e2562,
		0xc040b340,
		0x265e5a51,
		0xe9b6c7aa,
		0xd62f105d,
		0x2441453,
		0xd8a1e681,
		0xe7d3fbc8,
		0x21e1cde6,
		0xc33707d6,
		0xf4d50d87,
		0x455a14ed,
		0xa9e3e905,
		0xfcefa3f8,
		0x676f02d9,
		0x8d2a4c8a,
	},
	Table3: []uint32{
		// round3
		0xfffa3942,
		0x8771f681,
		0x6d9d6122,
		0xfde5380c,
		0xa4beea44,
		0x4bdecfa9,
		0xf6bb4b60,
		0xbebfbc70,
		0x289b7ec6,
		0xeaa127fa,
		0xd4ef3085,
		0x4881d05,
		0xd9d4d039,
		0xe6db99e5,
		0x1fa27cf8,
		0xc4ac5665,
	},
	Table4: []uint32{
		// round 4
		0xf4292244,
		0x432aff97,
		0xab9423a7,
		0xfc93a039,
		0x655b59c3,
		0x8f0ccc92,
		0xffeff47d,
		0x85845dd1,
		0x6fa87e4f,
		0xfe2ce6e0,
		0xa3014314,
		0x4e0811a1,
		0xf7537e82,
		0xbd3af235,
		0x2ad7d2bb,
		0xeb86d391,
	},
}

var program = `// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by go run gen.go -output md5block.go; DO NOT EDIT.

package md5

import (
	"encoding/binary"
	"math/bits"
)

func blockGeneric(dig *digest, p []byte) {
	// load state
	a, b, c, d := dig.s[0], dig.s[1], dig.s[2], dig.s[3]

	for i := 0; i <= len(p)-BlockSize; i += BlockSize {
		// eliminate bounds checks on p
		q := p[i:]
		q = q[:BlockSize:BlockSize]

		// save current state
		aa, bb, cc, dd := a, b, c, d

		// load input block
		{{range $i := seq 16 -}}
			{{printf "x%x := binary.LittleEndian.Uint32(q[4*%#x:])" $i $i}}
		{{end}}

		// round 1
		{{range $i, $s := dup 4 .Shift1 -}}
			{{printf "arg0 = arg1 + bits.RotateLeft32((((arg2^arg3)&arg1)^arg3)+arg0+x%x+%#08x, %d)" (idx 1 $i) (index $.Table1 $i) $s | relabel}}
			{{rotate -}}
		{{end}}
	
		// round 2
		{{range $i, $s := dup 4 .Shift2 -}}
			{{printf "arg0 = arg1 + bits.RotateLeft32((((arg1^arg2)&arg3)^arg2)+arg0+x%x+%#08x, %d)" (idx 2 $i) (index $.Table2 $i) $s | relabel}}
			{{rotate -}}
		{{end}}
	
		// round 3
		{{range $i, $s := dup 4 .Shift3 -}}
			{{printf "arg0 = arg1 + bits.RotateLeft32((arg1^arg2^arg3)+arg0+x%x+%#08x, %d)" (idx 3 $i) (index $.Table3 $i) $s | relabel}}
			{{rotate -}}
		{{end}}
	
		// round 4
		{{range $i, $s := dup 4 .Shift4 -}}
			{{printf "arg0 = arg1 + bits.RotateLeft32((arg2^(arg1|^arg3))+arg0+x%x+%#08x, %d)" (idx 4 $i) (index $.Table4 $i) $s | relabel}}
			{{rotate -}}
		{{end}}

		// add saved state
		a += aa
		b += bb
		c += cc
		d += dd
	}

	// save state
	dig.s[0], dig.s[1], dig.s[2], dig.s[3] = a, b, c, d
}
`

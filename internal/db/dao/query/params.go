package query

import (
	"strconv"
)

type params struct {
	count int
}

func newParams() params {
	return params{count: 1}
}

func (p *params) next() string {
	val := "$" + strconv.Itoa(p.count)
	p.count++
	return val
}

func (p *params) in(arr []int, args *[]any) string {
	str := "("
	for i, val := range arr {
		str += p.next()
		if i < len(arr)-1 {
			str += ", "
		}
		*args = append(*args, val)
	}
	str += ")"
	return str
}

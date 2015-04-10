package main

type Stack interface {
	Len() int
	Push(v interface{})
	Pop() interface{}
	ToSlice() []interface{}
}

type slice []interface{}

func NewStack() Stack {
	return new(slice)
}

func (s slice) Len() int {
	return len(s)
}

func (ps *slice) Push(v interface{}) {
	*ps = append(*ps, v)
}

func (ps *slice) Pop() interface{} {
	if len(*ps) == 0 {
		panic("tried to pop an empty stack")
	}
	v := (*ps)[len(*ps)-1]
	*ps = (*ps)[:len(*ps)-1]
	return v
}

func (s slice) ToSlice() []interface{} {
	for i, v := range s {
		switch v.(type) {
		case Stack:
			s[i] = v.(Stack).ToSlice()
		default:
			continue
		}
	}
	return s
}

package stack

type Stack struct {
	arr []*int32
}

func New() *Stack {
	return &Stack{
		arr: []*int32{},
	}
}

func (stack *Stack) Size() int {
	return len(stack.arr)
}

func (stack *Stack) IsEmpty() bool {
	return len(stack.arr) == 0
}

func (stack *Stack) Push(i int32) {
	stack.arr = append(stack.arr, &i)
}

func (stack *Stack) Pop() *int32 {
	l := len(stack.arr)
	if l == 0 {
		return nil
	}

	last := stack.arr[l-1]
	stack.arr = stack.arr[:l-1]
	return last
}

func (stack *Stack) Peek() *int32 {
	l := len(stack.arr)
	if l == 0 {
		return nil
	}

	return stack.arr[l-1]
}

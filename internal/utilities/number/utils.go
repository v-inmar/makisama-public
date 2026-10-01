package number_utils

func MakeSliceFromInt(n int) []int {
	s := make([]int, n)

	for i := range s {
		s[i] = i + 1
	}

	return s
}

func Addition(n int) int {
	return n + 1
}

func Subtraction(n int) int {
	return n - 1
}

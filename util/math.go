package util

type mathUtil struct{}

var Math mathUtil

func (mathUtil) AbsInt64(x int64) int64 {
	if x < 0 {
		x = -x
	}
	return x
}

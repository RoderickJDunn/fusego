package util

func MinInt(one int, two int) int {
	if one < two {
		return one
	}
	return two
}

func MaxInt(one int, two int) int {
	if one > two {
		return one
	}
	return two
}

func AbsInt(value int) int {
	if value < 0 {
		value *= -1
	}
	return value
}

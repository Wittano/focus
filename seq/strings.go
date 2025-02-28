package seq

func SumStringLength(s []string) (l int) {
	for _, v := range s {
		l += len(v)
	}

	return
}

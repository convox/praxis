package local

func coalesce(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}

	return ""
}

func coalescei(ints ...int) int {
	for _, i := range ints {
		if i > 0 {
			return i
		}
	}

	return 0
}

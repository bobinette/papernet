package web

func isStringIn(str string, a []string) bool {
	for _, e := range a {
		if e == str {
			return true
		}
	}
	return false
}

package util

func IsStrInList(l []string, i string) bool {
	for _, item := range l {
		if item == i {
			return true
		}
	}
	return false

}

func IsInt64InList(l []*int64, i int64) bool {
	for _, item := range l {
		if *item == i {
			return true
		}
	}
	return false

}

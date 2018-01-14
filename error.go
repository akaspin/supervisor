package supervisor

type MultiError []error

func (e MultiError) Error() (res string) {
	for i, e1 := range e {
		if i > 0 {
			res += ","
		}
		res += e1.Error()
	}
	return
}

// Append
func AppendError(left, right error) (err error) {
	if left == nil && right == nil {
		return
	}
	if left == nil {
		err = right
		return
	}
	if right == nil {
		err = left
		return
	}
	var err1 MultiError
	if l1, ok := left.(MultiError); ok {
		err1 = append(err1, l1...)
	} else {
		err1 = append(err1, left)
	}
	if l1, ok := right.(MultiError); ok {
		err1 = append(err1, l1...)
	} else {
		err1 = append(err1, right)
	}
	err = err1
	return
}

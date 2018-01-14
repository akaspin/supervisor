package supervisor

type MultiError struct {
	Errs []error
}

func (e *MultiError) Error() (res string) {
	for i, e1 := range e.Errs {
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
	err1 := &MultiError{}
	if l1, ok := left.(*MultiError); ok {
		err1.Errs = append(err1.Errs, l1.Errs...)
	} else {
		err1.Errs = append(err1.Errs, left)
	}
	if l1, ok := right.(*MultiError); ok {
		err1.Errs = append(err1.Errs, l1.Errs...)
	} else {
		err1.Errs = append(err1.Errs, right)
	}

	return
}

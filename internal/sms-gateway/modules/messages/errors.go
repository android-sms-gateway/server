package messages

type ErrValidation string

func (e ErrValidation) Error() string {
	return string(e)
}

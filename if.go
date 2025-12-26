package utils

// If returns ifTrue if cond is true, otherwise returns ifFalse
func If[T any](cond bool, ifTrue T, ifFalse T) T {
	if cond {
		return ifTrue
	}
	return ifFalse
}

package testdata

const _mySecretInt = 1729

type _secretStruct struct {
	_secretField int
	sharedField  int
}

var sharedInstanceOfSecretStruct = _secretStruct{123, 456}

func _getSecretInt() int {
	return _mySecretInt
}

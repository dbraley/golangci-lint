//args: -Eunderscorevisibility
package testdata

func UnderscoreVisibility() {
	_ = _getSecretInt() // ERROR "cannot refer to file-private _getSecretInt"

	_ = _mySecretInt // ERROR "cannot refer to file-private `_mySecretInt`"

	_ = _secretStruct{} // ERROR "cannot refer to file-private `_secretStruct`"

	ok := sharedInstanceOfSecretStruct

	_ = ok._secretField // TODO "cannot refer to file-private `_secretField`"
}

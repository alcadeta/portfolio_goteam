package login

// fakeReqValidator is a test fake for ReqValidator.
type fakeReqValidator struct {
	inReqBody ReqBody
	outOK     bool
}

// Validate implements the ReqValidator interface on fakeReqValidator. It
// assigns the parameters passed into it to their corresponding In... fields on
// the fake instance and returns its Out.. fields as per function signature.
func (f *fakeReqValidator) Validate(reqBody ReqBody) bool {
	f.inReqBody = reqBody
	return f.outOK
}

// fakeHashComparer is a test fake for Comparer.
type fakeHashComparer struct {
	inHash      []byte
	inPlaintext string
	outErr      error
}

// Compare implements the Comparer interface on fakeHashComparer.
func (f *fakeHashComparer) Compare(hash []byte, plaintext string) error {
	f.inHash, f.inPlaintext = hash, plaintext
	return f.outErr
}
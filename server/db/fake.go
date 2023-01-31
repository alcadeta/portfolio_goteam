package db

// FakeCloser is a test fake for Closer.
type FakeCloser struct{ IsCalled bool }

// Close implements the Closer interface on FakeCloser. It assigns the
// parameters passed into it to their corresponding In... fields on the fake
// instance.
func (c *FakeCloser) Close() { c.IsCalled = true }

// FakeUserInserter is a test fake for Inserter[User].
type FakeUserInserter struct {
	InUser User
	OutErr error
}

// Insert implements the Inserter[User] interface on FakeUserInserter. It
// assigns the parameters passed into it to their corresponding In... fields on
// the fake instance and returns its Out.. fields as per function signature.

func (f *FakeUserInserter) Insert(user User) error {
	f.InUser = user
	return f.OutErr
}

// FakeUserSelector is a test fake for Selector[User].
type FakeUserSelector struct {
	InUserID string
	OutRes   User
	OutErr   error
}

// Select implements the Selector[User] interface on FakeUserSelector. It
// assigns the parameters passed into it to their corresponding In... fields on
// the fake instance and returns its Out.. fields as per function signature.
func (f *FakeUserSelector) Select(userID string) (User, error) {
	f.InUserID = userID
	return f.OutRes, f.OutErr
}

// FakeCounter is a test fake for Counter.
type FakeCounter struct {
	InID   string
	OutRes int
	OutErr error
}

// Count implements the Counter interface on FakeCounter. It assigns the
// parameters passed into it to their corresponding In... fields on the fake
// instance and returns its Out.. fields as per function signature.
func (f *FakeCounter) Count(id string) (int, error) {
	f.InID = id
	return f.OutRes, f.OutErr
}

// FakeBoardInserter is a test fake for Inserter[Board].
type FakeBoardInserter struct {
	InBoard Board
	OutErr  error
}

// Insert implements the Inserter[Board] interface on FakeBoardInserter. It
// assigns the parameters passed into it to their corresponding In... fields on
// the fake instance and returns its Out.. fields as per function signature.
func (f *FakeBoardInserter) Insert(board Board) error {
	f.InBoard = board
	return f.OutErr
}

// FakeRelSelector is a test fake for RelSelector.
type FakeRelSelector struct {
	InIDA      string
	InIDB      string
	OutIsAdmin bool
	OutErr     error
}

// Select implements the RelSelector interface on FakeRelSelector. It assigns
// the parameters passed into it to their corresponding In... fields on the fake
// instance and returns its Out.. fields as per function signature.
func (f *FakeRelSelector) Select(idA, idB string) (bool, error) {
	f.InIDA, f.InIDB = idA, idB
	return f.OutIsAdmin, f.OutErr
}

// FakeDeleter is a test fake for Deleter.
type FakeDeleter struct {
	InID   string
	OutErr error
}

// Delete implements the Deleter interface on FakeDeleter. It assigns the
// parameters passed into it to their corresponding In... fields on the fake
// instance and returns its Out.. fields as per function signature.
func (f *FakeDeleter) Delete(id string) error {
	f.InID = id
	return f.OutErr
}
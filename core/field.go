package core

type Field struct {
	Key   string
	Value any
}

// F is a helper function to create a Field object.
func F(name string, value any) Field {
	return Field{
		Key:   name,
		Value: value,
	}
}

// E is a helper function to create a Field object for errors with key error.
func E(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}

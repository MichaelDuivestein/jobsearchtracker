package testutil

func ToPtr[Type any](variable Type) *Type {
	return &variable
}

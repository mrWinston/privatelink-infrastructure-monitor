package aws

func NewBoolPtr(b bool) *bool {
	tmpVar := b
	return &tmpVar
}
func NewInt32Ptr(b int32) *int32 {
	tmpVar := b
	return &tmpVar
}
func NewStringPtr(s string) *string {
	tmpVar := s
	return &tmpVar
}

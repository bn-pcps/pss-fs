package utils

// GetStringPtr returns a pointer to the string, or nil if string is empty
func GetStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

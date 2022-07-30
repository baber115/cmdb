package utils

// 避免空指针
func PtrStrV(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

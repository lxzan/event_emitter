package helper

func Uniq[T comparable](arr []T) []T {
	var m = make(map[T]struct{}, len(arr))
	var list = make([]T, 0, len(arr))
	for _, item := range arr {
		m[item] = struct{}{}
	}
	for k, _ := range m {
		list = append(list, k)
	}
	return list
}

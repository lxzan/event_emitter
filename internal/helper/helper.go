package helper

import "strings"

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

// Split 分割字符串(空值将会被过滤掉)
func Split(s string, sep string) []string {
	var list = strings.Split(s, sep)
	var j = 0
	for _, v := range list {
		if v = strings.TrimSpace(v); v != "" {
			list[j] = v
			j++
		}
	}
	return list[:j]
}

func Clone[T any, A ~[]T](a A) A {
	return append(A(nil), a...)
}

// Filter 过滤器
func Filter[T any, A ~[]T](arr A, check func(i int, v T) bool) A {
	var results = make([]T, 0, len(arr))
	for i, v := range arr {
		if check(i, v) {
			results = append(results, arr[i])
		}
	}
	return results
}

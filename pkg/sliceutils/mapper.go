package sliceutils

type MapFuncWithErr[T any, U any] func(T) (U, error)

func MapSliceWithErr[T any, U any](slice []T, fn MapFuncWithErr[T, U]) ([]U, error) {
	mapped := make([]U, len(slice))
	for i, item := range slice {
		var err error
		mapped[i], err = fn(item)
		if err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

type MapFunc[T any, U any] func(T) U

func MapSlice[T any, U any](slice []T, fn MapFunc[T, U]) []U {
	mapped := make([]U, len(slice))
	for i, item := range slice {
		mapped[i] = fn(item)
	}
	return mapped
}

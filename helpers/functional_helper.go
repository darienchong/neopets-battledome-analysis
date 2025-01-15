package helpers

import (
	"fmt"
	"sort"
)

type Tuple struct {
	Elements []any
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func MapMultiple[T any](ts []T, fn func(T) []any) []Tuple {
	result := make([]Tuple, len(ts))
	for i, t := range ts {
		result[i] = Tuple{Elements: fn(t)}
	}
	return result
}

func OrderBy[T any, V string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ts []T, keyFn func(T) V) []T {
	tsWithSortKey := MapMultiple(ts, func(t T) []any {
		return []any{t, keyFn(t)}
	})
	sort.Slice(tsWithSortKey, func(i, j int) bool {
		return tsWithSortKey[i].Elements[1].(V) < tsWithSortKey[j].Elements[1].(V)
	})
	return Map(tsWithSortKey, func(tuple Tuple) T {
		return tuple.Elements[0].(T)
	})
}

func OrderByDescending[T any, V string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ts []T, keyFn func(T) V) []T {
	tsWithSortKey := MapMultiple(ts, func(t T) []any {
		return []any{t, keyFn(t)}
	})
	sort.Slice(tsWithSortKey, func(i, j int) bool {
		return tsWithSortKey[i].Elements[1].(V) > tsWithSortKey[j].Elements[1].(V)
	})
	return Map(tsWithSortKey, func(tuple Tuple) T {
		return tuple.Elements[0].(T)
	})
}

func Filter[T any](ts []T, predicate func(T) bool) []T {
	filteredTs := []T{}
	for _, elt := range ts {
		if predicate(elt) {
			filteredTs = append(filteredTs, elt)
		}
	}
	return filteredTs
}

func FilterMap[K comparable, V any](m map[K]V, predicate func(V) bool) map[K]V {
	filteredMap := map[K]V{}

	for k, v := range m {
		if predicate(v) {
			filteredMap[k] = v
		}
	}
	return filteredMap
}

func FilterPointers[T any](ts []*T, predicate func(*T) bool) []*T {
	filteredTs := []*T{}
	for _, elt := range ts {
		if predicate(elt) {
			filteredTs = append(filteredTs, elt)
		}
	}
	return filteredTs
}

func Count[T any](ts []T, predicate func(T) bool) int {
	return len(Filter(ts, predicate))
}

func Sum[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](ts []T) T {
	sum := T(0)
	for _, v := range ts {
		sum += v
	}
	return sum
}

func Distinct[T comparable](ts []T) []T {
	tSet := map[T]bool{}
	distinctTs := []T{}
	for _, v := range ts {
		_, ok := tSet[v]
		if !ok {
			tSet[v] = true
			distinctTs = append(distinctTs, v)
		}
	}
	return distinctTs
}

func GroupBy[T any, K comparable](ts []T, keyFn func(T) K) map[K][]T {
	groups := map[K][]T{}
	for _, t := range ts {
		key := keyFn(t)
		_, ok := groups[key]
		if !ok {
			groups[key] = []T{}
		}
		groups[key] = append(groups[key], t)
	}
	return groups
}

func GroupPointersBy[T any, K comparable](ts []*T, keyFn func(*T) K) map[K][]*T {
	groups := map[K][]*T{}
	for _, t := range ts {
		key := keyFn(t)
		_, ok := groups[key]
		if !ok {
			groups[key] = []*T{}
		}
		groups[key] = append(groups[key], t)
	}
	return groups
}

func ToSlice[K comparable, V any](enumerable map[K]V) []Tuple {
	tuples := []Tuple{}
	for k, v := range enumerable {
		tuples = append(tuples, Tuple{Elements: []any{k, v}})
	}
	return tuples
}

func First[T any](ts []T, predicate func(T) bool) (T, error) {
	for _, t := range ts {
		if predicate(t) {
			return t, nil
		}
	}

	var zero T
	return zero, fmt.Errorf("no element matching predicate in given array")
}

func Reduce[T any](ts []T, reducer func(t1 T, t2 T) T) T {
	base := ts[0]
	for i, val := range ts {
		if i == 0 {
			continue
		}

		base = reducer(base, val)
	}
	return base
}

func ToMap[K comparable, T, V any](ts []T, keyFn func(T) K, valFn func(T) V) map[K]V {
	mappedVals := map[K]V{}
	for _, t := range ts {
		key := keyFn(t)
		val := valFn(t)
		mappedVals[key] = val
	}
	return mappedVals
}

func ToPointerMap[K comparable, T, V any](ts []T, keyFn func(T) K, valFn func(T) *V) map[K]*V {
	mappedVals := map[K]*V{}
	for _, t := range ts {
		key := keyFn(t)
		val := valFn(t)
		mappedVals[key] = val
	}
	return mappedVals
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := []K{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func PointerValues[K comparable, V any](m map[K]*V) []*V {
	values := []*V{}
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func Values[K comparable, V any](m map[K]V) []V {
	values := []V{}
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func LazyWhen[T any](pred bool, ifTrue func() T, ifFalse func() T) T {
	if pred {
		return ifTrue()
	}
	return ifFalse()
}

func When[T any](pred bool, ifTrue T, ifFalse T) T {
	if pred {
		return ifTrue
	}
	return ifFalse
}

func AsLiteral[T any](ptrs []*T) []T {
	return Map(ptrs, func(ptr *T) T {
		return *ptr
	})
}

func ReduceMap[K comparable, V any](m map[K][]*V, reducer func(*V, *V) *V) map[K]*V {
	m2 := map[K]*V{}
	for k, vs := range m {
		m2[k] = Reduce(vs, reducer)
	}
	return m2
}

func Max[T comparable](ts []T, less func(T, T) bool) T {
	best := ts[0]
	for _, t := range ts {
		if less(best, t) {
			best = t
		}
	}
	return best
}

func FlatMap[T, V any](ts []T, arrayFn func(T) []V) []V {
	flattened := []V{}
	for _, t := range ts {
		flattened = append(flattened, arrayFn(t)...)
	}
	return flattened
}

func FlatMapPointer[T, V any](ts []T, arrayFn func(T) []*V) []*V {
	flattened := []*V{}
	for _, t := range ts {
		flattened = append(flattened, arrayFn(t)...)
	}
	return flattened
}

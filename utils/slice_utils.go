package utils

func MapSlice[T any, M any](a []T, f func(T) M) []M {
  n := make([]M, len(a))
  for i, e := range a {
    n[i] = f(e)
  }
  return n
}

func Find(slice []string, val string) (int, bool) {
  for i, item := range slice {
    if item == val {
      return i, true
    }
  }
  return -1, false
}

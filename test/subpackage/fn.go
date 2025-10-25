package subpackage

type Test struct {
	Name string
}

func (s *Test) SetName(name string) {
	s.Name = name
}

func (s *Test) GetName() string {
	return s.Name
}

//go:generate testifai --func=Sum --type=table --output=sum_table_ai_test.go
func Sum(a, b int) int {
	if a > b {
		return a
	}
	if a == b {
		return a + b
	}
	return b
}

func a() int {
	return 1
}

func b() int {
	return 2
}

func c() int {
	return 3
}

//go:generate testifai --func=D --type=table --output=D_ai_test.go
func D(x int) int {
	return x + a() + b() + c()
}

package subpackage

type Test struct {
	Name string
}

//go:generate testifai --type=xunit --output=other_test.go

func (s *Test) SetName(name string) {
	s.Name = name
}

//go:generate testifai --func=GetName --type=suite --output=get_name_suite_ai_test.go
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

func D() int {
	return a() + b() + c()
}

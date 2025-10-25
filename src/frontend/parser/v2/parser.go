package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

// --- 1. Итоговые и вспомогательные структуры ---

// CollectedFunction - результат сбора для одной функции или метода.
type CollectedFunction struct {
	FunctionName         string   `json:"functionName"`
	PackageName          string   `json:"packageName"`
	ExternalImports      []string `json:"externalImports"`
	BodyWithReceiver     string   `json:"bodyWithReceiver"`
	InternalDependencies string   `json:"internalDependencies"`
}

// Declaration представляет одно объявление в коде.
type Declaration struct {
	Key         string // Уникальный ключ: "packageName.DeclName" или "packageName.(Receiver).MethodName"
	Name        string
	Receiver    string // Для методов: имя структуры-ресивера ("User")
	PackageName string
	Filepath    string
	Source      string
	Node        ast.Node
}

// --- 2. Глобальный парсер контекста репозитория ---

// RepositoryContext хранит всю информацию о коде в репозитории.
type RepositoryContext struct {
	ModulePath   string
	RepoRoot     string
	Declarations map[string]*Declaration
	FileImports  map[string][]*ast.ImportSpec // [filepath] -> imports
	FileSet      *token.FileSet
	FileSources  map[string][]byte
}

func NewRepositoryContext(repoRoot string) (*RepositoryContext, error) {
	// Читаем go.mod, чтобы узнать путь модуля
	modPath, err := getModulePath(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("не удалось определить модуль go: %w", err)
	}

	return &RepositoryContext{
		ModulePath:   modPath,
		RepoRoot:     repoRoot,
		Declarations: make(map[string]*Declaration),
		FileImports:  make(map[string][]*ast.ImportSpec),
		FileSet:      token.NewFileSet(),
		FileSources:  make(map[string][]byte),
	}, nil
}

// Parse анализирует все .go файлы.
func (c *RepositoryContext) Parse() error {
	return filepath.WalkDir(c.RepoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			c.FileSources[path] = src

			f, err := parser.ParseFile(c.FileSet, path, src, parser.ParseComments)
			if err != nil {
				return err
			}

			c.FileImports[path] = f.Imports
			packageName := f.Name.Name

			for _, decl := range f.Decls {
				c.addDeclaration(decl, packageName, path)
			}
		}
		return nil
	})
}

func (c *RepositoryContext) addDeclaration(decl ast.Decl, packageName, path string) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		name := d.Name.Name
		key := fmt.Sprintf("%s.%s", packageName, name)
		receiver := ""

		if d.Recv != nil && len(d.Recv.List) > 0 {
			recvType := d.Recv.List[0].Type
			recvName := ""
			if starExpr, ok := recvType.(*ast.StarExpr); ok {
				if ident, ok := starExpr.X.(*ast.Ident); ok {
					recvName = ident.Name
				}
			} else if ident, ok := recvType.(*ast.Ident); ok {
				recvName = ident.Name
			}
			if recvName != "" {
				receiver = recvName
				key = fmt.Sprintf("%s.(%s).%s", packageName, recvName, name)
			}
		}

		c.Declarations[key] = &Declaration{
			Key:         key,
			Name:        name,
			Receiver:    receiver,
			PackageName: packageName,
			Filepath:    path,
			Source:      c.getSource(path, d),
			Node:        d,
		}

	case *ast.GenDecl:
		for _, spec := range d.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				key := fmt.Sprintf("%s.%s", packageName, ts.Name.Name)
				c.Declarations[key] = &Declaration{
					Key:         key,
					Name:        ts.Name.Name,
					PackageName: packageName,
					Filepath:    path,
					Source:      c.getSource(path, d),
					Node:        d,
				}
			}
		}
	}
}

func (c *RepositoryContext) getSource(filepath string, node ast.Node) string {
	src := c.FileSources[filepath]
	start := c.FileSet.Position(node.Pos()).Offset
	end := c.FileSet.Position(node.End()).Offset
	if start >= 0 && end >= 0 && end <= len(src) {
		return string(src[start:end])
	}
	return ""
}

// --- 3. Основная логика сбора ---

// Collect запускает процесс сбора зависимостей.
func Collect(ctx *RepositoryContext, targetFilePath string, targetFuncName string) ([]CollectedFunction, error) {
	// Находим AST нужного файла, чтобы определить пакет и список функций
	f, err := parser.ParseFile(ctx.FileSet, targetFilePath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти AST для файла %s: %w", targetFilePath, err)
	}
	packageName := f.Name.Name

	var targets []*Declaration
	if targetFuncName != "" {
		// Ищем одну конкретную функцию/метод
		// Примечание: для простоты ищем по имени, для методов может потребоваться более сложная логика
		key := fmt.Sprintf("%s.%s", packageName, targetFuncName) // Упрощенный поиск
		if decl, ok := ctx.Declarations[key]; ok {
			targets = append(targets, decl)
		} else {
			// Попробуем поискать как метод (перебирая все декларации)
			for _, d := range ctx.Declarations {
				if d.PackageName == packageName && d.Name == targetFuncName && d.Receiver != "" {
					targets = append(targets, d)
					break
				}
			}
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("функция или метод '%s' не найдены в файле %s", targetFuncName, targetFilePath)
		}

	} else {
		// Собираем все функции и методы из указанного файла
		for _, decl := range ctx.Declarations {
			if decl.Filepath == targetFilePath && decl.Node.Pos().IsValid() {
				if _, ok := decl.Node.(*ast.FuncDecl); ok {
					targets = append(targets, decl)
				}
			}
		}
	}

	var results []CollectedFunction
	for _, target := range targets {
		collector := &dependencyCollector{
			ctx:             ctx,
			startDecl:       target,
			internalDeps:    make(map[string]*Declaration),
			externalImports: make(map[string]bool),
			visited:         make(map[string]bool),
		}
		collector.collect(target)

		// Собираем результат
		result, err := collector.assembleResult()
		if err != nil {
			return nil, fmt.Errorf("ошибка сборки результата для %s: %w", target.Name, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// dependencyCollector собирает зависимости для ОДНОЙ функции/метода.
type dependencyCollector struct {
	ctx             *RepositoryContext
	startDecl       *Declaration
	internalDeps    map[string]*Declaration
	externalImports map[string]bool // set of import paths
	visited         map[string]bool
}

func (c *dependencyCollector) collect(decl *Declaration) {
	if c.visited[decl.Key] {
		return
	}
	c.visited[decl.Key] = true

	// Не добавляем стартовую функцию в ее собственные зависимости
	if decl.Key != c.startDecl.Key {
		c.internalDeps[decl.Key] = decl
	}

	ast.Inspect(decl.Node, func(node ast.Node) bool {
		switch n := node.(type) {
		// Обработка вызовов вида pkg.Function()
		case *ast.SelectorExpr:
			if ident, ok := n.X.(*ast.Ident); ok {
				pkgName := ident.Name
				// Ищем, что это за пакет
				for _, imp := range c.ctx.FileImports[decl.Filepath] {
					// Имя пакета может быть псевдонимом (import f "fmt")
					var currentPkgName string
					if imp.Name != nil {
						currentPkgName = imp.Name.Name
					} else {
						// Имя из пути импорта
						parts := strings.Split(strings.Trim(imp.Path.Value, `"`), "/")
						currentPkgName = parts[len(parts)-1]
					}

					if pkgName == currentPkgName {
						importPath := strings.Trim(imp.Path.Value, `"`)
						// Если импорт не из нашего репозитория - это внешняя зависимость
						if !strings.HasPrefix(importPath, c.ctx.ModulePath) {
							c.externalImports[importPath] = true
						} else {
							// TODO: Обработка зависимостей из других пакетов этого же репозитория
						}
						return false // Глубже не идем, т.к. нашли пакет
					}
				}
			}

		// Обработка вызовов вида Function() или использования типа Type
		case *ast.Ident:
			// Ищем объявление с таким именем в том же пакете
			key := fmt.Sprintf("%s.%s", decl.PackageName, n.Name)
			if depDecl, ok := c.ctx.Declarations[key]; ok {
				c.collect(depDecl)
			}
			// Также ищем методы для этого типа
			if depDecl, ok := c.ctx.Declarations[fmt.Sprintf("%s.(%s)", decl.PackageName, n.Name)]; ok {
				// Этот кейс сложнее, может привести к рекурсии. Пока пропустим.
				_ = depDecl
			}
		}
		return true
	})
}

func (c *dependencyCollector) assembleResult() (*CollectedFunction, error) {
	// Собираем тело
	body := c.startDecl.Source
	if c.startDecl.Receiver != "" {
		// Если это метод, находим объявление его ресивера
		recvKey := fmt.Sprintf("%s.%s", c.startDecl.PackageName, c.startDecl.Receiver)
		if recvDecl, ok := c.ctx.Declarations[recvKey]; ok {
			body = recvDecl.Source + "\n\n" + body
		}
	}

	// Собираем внутренние зависимости
	var internalDepsBuilder strings.Builder
	// Сортируем для детерминированного вывода
	keys := make([]string, 0, len(c.internalDeps))
	for k := range c.internalDeps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		internalDepsBuilder.WriteString(c.internalDeps[k].Source)
		internalDepsBuilder.WriteString("\n\n")
	}

	// Собираем внешние импорты
	extImports := make([]string, 0, len(c.externalImports))
	for imp := range c.externalImports {
		extImports = append(extImports, imp)
	}
	sort.Strings(extImports)

	return &CollectedFunction{
		FunctionName:         c.startDecl.Name,
		PackageName:          c.startDecl.PackageName,
		ExternalImports:      extImports,
		BodyWithReceiver:     body,
		InternalDependencies: internalDepsBuilder.String(),
	}, nil
}

func getModulePath(repoRoot string) (string, error) {
	goModPath := filepath.Join(repoRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(data), nil
}

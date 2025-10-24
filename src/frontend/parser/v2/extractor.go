package v2

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type MethodInfo struct {
	ReceiverType string
	MethodName   string
	Decl         *ast.FuncDecl
}

type FunctionExtractor struct {
	fset         *token.FileSet
	files        []*ast.File
	targetFunc   string
	targetType   string // For methods: the receiver type
	targetMethod string // For methods: the method name
	isMethod     bool
	imports      map[string]string
	functions    map[string]*ast.FuncDecl
	methods      map[string]map[string]*ast.FuncDecl // receiverType -> methodName -> FuncDecl
	types        map[string]ast.Decl
	constants    map[string]*ast.GenDecl
	variables    map[string]*ast.GenDecl
	usedIdents   map[string]bool
	usedMethods  map[string]map[string]bool // receiverType -> methodName -> bool (tracks actually called methods)
	packageName  string
}

func NewFunctionExtractorFromFile(filePath string) (*FunctionExtractor, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	extractor := &FunctionExtractor{
		fset:        fset,
		files:       []*ast.File{file},
		imports:     make(map[string]string),
		functions:   make(map[string]*ast.FuncDecl),
		methods:     make(map[string]map[string]*ast.FuncDecl),
		types:       make(map[string]ast.Decl),
		constants:   make(map[string]*ast.GenDecl),
		variables:   make(map[string]*ast.GenDecl),
		usedIdents:  make(map[string]bool),
		usedMethods: make(map[string]map[string]bool),
		packageName: file.Name.Name,
	}

	return extractor, nil
}

func NewFunctionExtractorFromDir(dirPath string) (*FunctionExtractor, error) {
	fset := token.NewFileSet()
	var allFiles []*ast.File
	var packageName string

	// Walk through directory and subdirectories
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse error in %s: %w", path, err)
		}

		if packageName == "" {
			packageName = file.Name.Name
		}

		allFiles = append(allFiles, file)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in %s", dirPath)
	}

	extractor := &FunctionExtractor{
		fset:        fset,
		files:       allFiles,
		imports:     make(map[string]string),
		functions:   make(map[string]*ast.FuncDecl),
		methods:     make(map[string]map[string]*ast.FuncDecl),
		types:       make(map[string]ast.Decl),
		constants:   make(map[string]*ast.GenDecl),
		variables:   make(map[string]*ast.GenDecl),
		usedIdents:  make(map[string]bool),
		usedMethods: make(map[string]map[string]bool),
		packageName: packageName,
	}

	return extractor, nil
}

// parseTarget parses target in format: Function or Type.Method or package.Type.Method
func (fe *FunctionExtractor) parseTarget(target string) {
	parts := strings.Split(target, ".")

	if len(parts) == 1 {
		// Simple function name
		fe.targetFunc = parts[0]
		fe.isMethod = false
	} else if len(parts) == 2 {
		// Type.Method
		fe.targetType = parts[0]
		fe.targetMethod = parts[1]
		fe.isMethod = true
	} else if len(parts) == 3 {
		// package.Type.Method (ignore package for now, use Type.Method)
		fe.targetType = parts[1]
		fe.targetMethod = parts[2]
		fe.isMethod = true
	}
}

func (fe *FunctionExtractor) Extract(targetFunction string) (string, error) {
	fe.parseTarget(targetFunction)

	fe.collectDeclarations()

	if fe.isMethod {
		return fe.extractMethod(targetFunction)
	}
	return fe.extractFunction(targetFunction)
}

func (fe *FunctionExtractor) extractFunction(targetFunction string) (string, error) {

	targetFuncDecl, ok := fe.functions[targetFunction]
	if !ok {
		return "", fmt.Errorf("function %s not found", fe.targetFunc)
	}

	fe.collectDependencies(targetFuncDecl)
	return fe.generateCode(targetFuncDecl)
}

func (fe *FunctionExtractor) extractMethod(targetMethod string) (string, error) {
	var methodDecl *ast.FuncDecl
	var foundType string

	// Try to find the method in different receiver type variants
	candidates := []string{fe.targetType, "*" + fe.targetType}

	// If already starts with *, also try without it
	if strings.HasPrefix(fe.targetType, "*") {
		candidates = append(candidates, strings.TrimPrefix(fe.targetType, "*"))
	}

	for _, candidate := range candidates {
		if typeMethods, ok := fe.methods[candidate]; ok {
			if decl, ok := typeMethods[targetMethod]; ok {
				methodDecl = decl
				foundType = candidate
				break
			}
		}
	}

	if methodDecl == nil {
		return "", fmt.Errorf("method %s not found for type %s (tried: %v)",
			targetMethod, fe.targetType, candidates)
	}

	fe.targetType = foundType

	// Mark the receiver type as used
	receiverTypeName := strings.TrimPrefix(fe.targetType, "*")
	fe.usedIdents[receiverTypeName] = true

	fe.collectDependencies(methodDecl)
	return fe.generateCodeForMethod(methodDecl)
}

func (fe *FunctionExtractor) collectDeclarations() {
	for _, file := range fe.files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					// Regular function
					fe.functions[d.Name.Name] = d
				} else {
					// Method
					receiverType := fe.getReceiverType(d.Recv)
					if fe.methods[receiverType] == nil {
						fe.methods[receiverType] = make(map[string]*ast.FuncDecl)
					}
					fe.methods[receiverType][d.Name.Name] = d
				}
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE:
					for _, spec := range d.Specs {
						typeSpec := spec.(*ast.TypeSpec)
						fe.types[typeSpec.Name.Name] = d
					}
				case token.CONST:
					for _, spec := range d.Specs {
						valueSpec := spec.(*ast.ValueSpec)
						for _, name := range valueSpec.Names {
							fe.constants[name.Name] = d
						}
					}
				case token.VAR:
					for _, spec := range d.Specs {
						valueSpec := spec.(*ast.ValueSpec)
						for _, name := range valueSpec.Names {
							fe.variables[name.Name] = d
						}
					}
				case token.IMPORT:
					for _, spec := range d.Specs {
						importSpec := spec.(*ast.ImportSpec)
						path := strings.Trim(importSpec.Path.Value, "\"")
						alias := ""
						if importSpec.Name != nil {
							alias = importSpec.Name.Name
						} else {
							parts := strings.Split(path, "/")
							alias = parts[len(parts)-1]
						}
						fe.imports[alias] = path
					}
				}
			}
		}
	}
}

func (fe *FunctionExtractor) getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	}
	return ""
}

func (fe *FunctionExtractor) collectDependencies(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			name := x.Name
			if fe.usedIdents[name] || isBuiltin(name) {
				return true
			}

			fe.usedIdents[name] = true

			// Only collect type definitions, constants and variables
			// DO NOT recursively collect functions and methods
			if typeDecl, ok := fe.types[name]; ok {
				fe.collectTypeDependencies(typeDecl)
			}
			if constDecl, ok := fe.constants[name]; ok {
				fe.collectDependencies(constDecl)
			}
			if varDecl, ok := fe.variables[name]; ok {
				fe.collectDependencies(varDecl)
			}
		case *ast.SelectorExpr:
			if ident, ok := x.X.(*ast.Ident); ok {
				fe.usedIdents[ident.Name] = true
				// Just mark the type as used, don't collect the method
			}
		}
		return true
	})
}

func (fe *FunctionExtractor) collectTypeDependencies(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			name := ident.Name
			if !fe.usedIdents[name] && !isBuiltin(name) {
				fe.usedIdents[name] = true
				if typeDecl, ok := fe.types[name]; ok {
					fe.collectTypeDependencies(typeDecl)
				}
			}
		}
		return true
	})
}

func isBuiltin(name string) bool {
	builtins := map[string]bool{
		"bool": true, "byte": true, "complex64": true, "complex128": true,
		"error": true, "float32": true, "float64": true, "int": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"rune": true, "string": true, "uint": true, "uint8": true,
		"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"true": true, "false": true, "iota": true, "nil": true,
		"append": true, "cap": true, "close": true, "complex": true,
		"copy": true, "delete": true, "imag": true, "len": true,
		"make": true, "new": true, "panic": true, "print": true,
		"println": true, "real": true, "recover": true,
	}
	return builtins[name]
}

func (fe *FunctionExtractor) generateCode(targetFunc *ast.FuncDecl) (string, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("package %s\n\n", fe.packageName))

	fe.writeImports(&builder)
	fe.writeTypes(&builder)
	fe.writeConstants(&builder)
	fe.writeVariables(&builder)
	// DO NOT write other functions - only the target function

	if err := fe.writeDecl(&builder, targetFunc); err != nil {
		return "", err
	}
	builder.WriteString("\n")

	return builder.String(), nil
}

func (fe *FunctionExtractor) generateCodeForMethod(targetMethod *ast.FuncDecl) (string, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("package %s\n\n", fe.packageName))

	fe.writeImports(&builder)
	fe.writeTypes(&builder)
	fe.writeConstants(&builder)
	fe.writeVariables(&builder)
	// DO NOT write other functions and methods - only the target method

	if err := fe.writeDecl(&builder, targetMethod); err != nil {
		return "", err
	}
	builder.WriteString("\n")

	return builder.String(), nil
}

func (fe *FunctionExtractor) writeImports(builder *strings.Builder) {
	usedImports := make(map[string]string)
	for ident := range fe.usedIdents {
		if path, ok := fe.imports[ident]; ok {
			usedImports[ident] = path
		}
	}

	if len(usedImports) > 0 {
		builder.WriteString("import (\n")
		for alias, path := range usedImports {
			parts := strings.Split(path, "/")
			pkgName := parts[len(parts)-1]
			if alias == pkgName {
				builder.WriteString(fmt.Sprintf("\t\"%s\"\n", path))
			} else {
				builder.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, path))
			}
		}
		builder.WriteString(")\n\n")
	}
}

func (fe *FunctionExtractor) writeTypes(builder *strings.Builder) {
	processedTypes := make(map[string]bool)
	for typeName := range fe.types {
		if fe.usedIdents[typeName] && !processedTypes[typeName] {
			processedTypes[typeName] = true
			if err := fe.writeDecl(builder, fe.types[typeName]); err != nil {
				continue
			}
			builder.WriteString("\n\n")
		}
	}
}

func (fe *FunctionExtractor) writeConstants(builder *strings.Builder) {
	processedConsts := make(map[string]bool)
	for constName := range fe.constants {
		constDecl := fe.constants[constName]
		if fe.usedIdents[constName] && !processedConsts[fmt.Sprintf("%p", constDecl)] {
			processedConsts[fmt.Sprintf("%p", constDecl)] = true
			if err := fe.writeDecl(builder, constDecl); err != nil {
				continue
			}
			builder.WriteString("\n\n")
		}
	}
}

func (fe *FunctionExtractor) writeVariables(builder *strings.Builder) {
	processedVars := make(map[string]bool)
	for varName := range fe.variables {
		varDecl := fe.variables[varName]
		if fe.usedIdents[varName] && !processedVars[fmt.Sprintf("%p", varDecl)] {
			processedVars[fmt.Sprintf("%p", varDecl)] = true
			if err := fe.writeDecl(builder, varDecl); err != nil {
				continue
			}
			builder.WriteString("\n\n")
		}
	}
}

func (fe *FunctionExtractor) writeFunctions(builder *strings.Builder, excludeFunc string) {
	for funcName, funcDecl := range fe.functions {
		if funcName != excludeFunc && fe.usedIdents[funcName] {
			if err := fe.writeDecl(builder, funcDecl); err != nil {
				continue
			}
			builder.WriteString("\n\n")
		}
	}
}

func (fe *FunctionExtractor) writeMethods(builder *strings.Builder, excludeType, excludeMethod string) {
	// Only write methods that are actually used
	for receiverType, usedMethodsForType := range fe.usedMethods {
		for methodName := range usedMethodsForType {
			// Skip the target method as it will be written separately
			if receiverType == excludeType && methodName == excludeMethod {
				continue
			}

			// Get the method declaration
			if typeMethods, ok := fe.methods[receiverType]; ok {
				if methodDecl, ok := typeMethods[methodName]; ok {
					if err := fe.writeDecl(builder, methodDecl); err != nil {
						continue
					}
					builder.WriteString("\n\n")
				}
			}
		}
	}
}

type ParseTarget struct {
	FuncName   *string // Full path for func name
	StructName *string
	TestType   string
	Output     *string
}

// func (fe *FunctionExtractor) FindAllDocumented() []ParseTarget {
// 	for _, function := range fe.functions {

// 	}

// 	for _, function := range fe.methods {

// 	}
// }

// func (fe *FunctionExtractor)

func (fe *FunctionExtractor) writeDecl(builder *strings.Builder, decl ast.Decl) error {
	return printer.Fprint(builder, fe.fset, decl)
}

// func main() {
// 	filePath := flag.String("file", "", "Path to Go file")
// 	dirPath := flag.String("dir", "", "Path to directory with Go project")
// 	target := flag.String("target", "", "Target to extract: Function or Type.Method")
// 	// Keep old -func flag for backward compatibility
// 	funcName := flag.String("func", "", "Function name to extract (deprecated, use -target)")
// 	output := flag.String("output", "", "Output file path (stdout if not specified)")

// 	flag.Parse()

// 	// Determine target
// 	finalTarget := *target
// 	if finalTarget == "" && *funcName != "" {
// 		finalTarget = *funcName
// 	}

// 	if finalTarget == "" {
// 		fmt.Println("Usage:")
// 		fmt.Println("  Extract function from single file:")
// 		fmt.Println("    test_extract -file <file_path> -target <function_name> [-output <output_file>]")
// 		fmt.Println("  Extract method from single file:")
// 		fmt.Println("    test_extract -file <file_path> -target <Type.Method> [-output <output_file>]")
// 		fmt.Println("  Extract from directory:")
// 		fmt.Println("    test_extract -dir <directory_path> -target <function_or_Type.Method> [-output <output_file>]")
// 		fmt.Println("  Extract with package qualifier:")
// 		fmt.Println("    test_extract -dir <directory_path> -target <package.Type.Method> [-output <output_file>]")
// 		fmt.Println()
// 		fmt.Println("Examples:")
// 		fmt.Println("  ./test_extract -file main.go -target ProcessData")
// 		fmt.Println("  ./test_extract -dir ./myapp -target User.Validate")
// 		fmt.Println("  ./test_extract -dir ./myapp -target myapp.User.Save")
// 		flag.PrintDefaults()
// 		os.Exit(1)
// 	}

// 	if *filePath == "" && *dirPath == "" {
// 		fmt.Println("Error: Either -file or -dir must be specified")
// 		flag.PrintDefaults()
// 		os.Exit(1)
// 	}

// 	if *filePath != "" && *dirPath != "" {
// 		fmt.Println("Error: Cannot use both -file and -dir at the same time")
// 		flag.PrintDefaults()
// 		os.Exit(1)
// 	}

// 	var extractor *FunctionExtractor
// 	var err error

// 	if *filePath != "" {
// 		extractor, err = NewFunctionExtractorFromFile(*filePath)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
// 			os.Exit(1)
// 		}
// 	} else {
// 		absDir, err := filepath.Abs(*dirPath)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
// 			os.Exit(1)
// 		}

// 		extractor, err = NewFunctionExtractorFromDir(absDir)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
// 			os.Exit(1)
// 		}
// 	}

// 	code, err := extractor.Extract("asd")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Extract error: %v\n", err)
// 		os.Exit(1)
// 	}

// 	if *output != "" {
// 		err = os.WriteFile(*output, []byte(code), 0644)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
// 			os.Exit(1)
// 		}
// 		fmt.Printf("Code successfully saved to %s\n", *output)
// 	} else {
// 		fmt.Println(code)
// 	}
// }

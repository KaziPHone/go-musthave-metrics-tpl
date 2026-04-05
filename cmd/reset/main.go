package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: reset <корневая_директория>")
		return
	}

	rootDir := os.Args[1]

	// Сканируем все пакеты, начиная с корневой директории
	packages := scanPackages(rootDir)

	// Обрабатываем каждый пакет
	for importPath, files := range packages {
		// Ищем все структуры с комментарием // generate:reset
		structs := findStructsWithComment(files)

		if len(structs) == 0 {
			continue
		}

		// Генерируем методы Reset() для всех структур в этом пакете
		var buf bytes.Buffer
		buf.WriteString("// Код сгенерирован утилитой reset. Не редактировать.\n")
		buf.WriteString("package " + filepath.Base(importPath) + "\n\n")

		for _, str := range structs {
			buf.WriteString(generateResetMethod(str))
			buf.WriteString("\n")
		}

		// Форматируем сгенерированный код
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Printf("Ошибка форматирования кода для %s: %v\n", importPath, err)
			return
		}

		// Записываем в reset.gen.go в директории пакета
		packageDir := filepath.Join(rootDir, filepath.FromSlash(importPath))
		outputFile := filepath.Join(packageDir, "reset.gen.go")

		if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
			fmt.Printf("Ошибка записи файла %s: %v\n", outputFile, err)
			return
		}

		fmt.Printf("Сгенерировано методов Reset() для %d структур в %s\n", len(structs), importPath)
	}
}

// scanPackages сканирует все пакеты, начиная с корневой директории
func scanPackages(rootDir string) map[string][]string {
	packages := make(map[string][]string)

	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем скрытые директории
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		// Пропускаем директории vendor, cmd и системные
		if d.IsDir() && (d.Name() == "vendor" || d.Name() == "cmd" || d.Name() == ".git" || d.Name() == ".pytest_cache" || d.Name() == ".idea" || d.Name() == "staticlint") {
			return fs.SkipDir
		}

		if !d.IsDir() {
			return nil
		}

		// Проверяем, содержит ли директория Go-пакет
		goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
		if err != nil || len(goFiles) == 0 {
			return nil
		}

		var packageFiles []string

		// Парсим каждый Go-файл для поиска имени пакета
		for _, file := range goFiles {
			if strings.HasSuffix(file, "_test.go") || strings.HasSuffix(file, "reset.gen.go") {
				continue
			}

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
			if err != nil {
				continue
			}

			if f.Name.Name == "" {
				continue
			}

			packageFiles = append(packageFiles, file)
		}

		if len(packageFiles) > 0 {
			// Вычисляем import path относительно корня
			relPath, _ := filepath.Rel(rootDir, path)
			importPath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

			packages[importPath] = packageFiles
		}

		return nil
	})

	return packages
}

// structInfo содержит информацию о структуре, найденной генератором
type structInfo struct {
	Name     string
	Fields   []fieldInfo
	RecvType string // *StructName или StructName
	Package  string
}

// fieldInfo содержит информацию о поле
type fieldInfo struct {
	Name      string
	TypeExpr  ast.Expr
	TypeName  string
	IsPointer bool
	IsSlice   bool
	IsMap     bool
	Comment   string
}

// findStructsWithComment находит все структуры с комментарием // generate:reset
func findStructsWithComment(filePaths []string) []structInfo {
	var result []structInfo

	for _, filePath := range filePaths {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Собираем типы из импортов для проверки типов
		config := &types.Config{Error: func(err error) {}}
		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}

		config.Check(f.Name.Name, fset, []*ast.File{f}, info)

		// Обрабатываем каждое объявление верхнего уровня
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			// Ищем объявления типов-структур
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Проверяем, является ли это структурой
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Ищем комментарий // generate:reset
				if !hasGenerateResetComment(typeSpec.Doc) && !hasGenerateResetComment(genDecl.Doc) {
					continue
				}

				// Собираем информацию о полях
				fields := collectFields(structType.Fields, info)

				// Определяем тип получателя (используем указатель для структур)
				recvType := "*" + typeSpec.Name.Name

				result = append(result, structInfo{
					Name:     typeSpec.Name.Name,
					Fields:   fields,
					RecvType: recvType,
					Package:  f.Name.Name,
				})
			}
		}
	}

	return result
}

// hasGenerateResetComment проверяет, содержит ли список комментариев // generate:reset
func hasGenerateResetComment(cl *ast.CommentGroup) bool {
	if cl == nil {
		return false
	}
	for _, comment := range cl.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if text == "generate:reset" {
			return true
		}
	}
	return false
}

// collectFields собирает информацию о полях структуры
func collectFields(fieldList *ast.FieldList, info *types.Info) []fieldInfo {
	var fields []fieldInfo

	if fieldList == nil {
		return fields
	}

	for _, field := range fieldList.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name
		if fieldName == "_" || strings.HasPrefix(fieldName, "_") {
			continue
		}

		// Получаем информацию о типе
		typeInfo := getTypeInfo(field.Type, info)
		comment := ""

		if field.Comment != nil && len(field.Comment.List) > 0 {
			comment = strings.TrimSpace(strings.TrimPrefix(field.Comment.List[0].Text, "//"))
		}

		fields = append(fields, fieldInfo{
			Name:      fieldName,
			TypeExpr:  field.Type,
			TypeName:  typeInfo.name,
			IsPointer: typeInfo.isPointer,
			IsSlice:   typeInfo.isSlice,
			IsMap:     typeInfo.isMap,
			Comment:   comment,
		})
	}

	return fields
}

// typeInfo содержит информацию о классификации типа
type typeInfo struct {
	name      string
	isPointer bool
	isSlice   bool
	isMap     bool
}

// getTypeInfo определяет характеристики типа
func getTypeInfo(expr ast.Expr, info *types.Info) typeInfo {
	ti := typeInfo{}

	switch t := expr.(type) {
	case *ast.StarExpr:
		ti.isPointer = true
		ti.name = extractTypeName(t.X, info)
	case *ast.Ident:
		ti.name = t.Name
	case *ast.SelectorExpr:
		ti.name = t.Sel.Name
	case *ast.ArrayType:
		ti.isSlice = true
		ti.name = "slice"
	case *ast.MapType:
		ti.isMap = true
		ti.name = "map"
	}

	return ti
}

// extractTypeName извлекает имя типа из выражения
func extractTypeName(expr ast.Expr, info *types.Info) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return extractTypeName(t.X, info)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

// generateResetMethod генерирует метод Reset() для структуры
func generateResetMethod(s structInfo) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("func (rs %s) Reset() {\n", s.RecvType))
	buf.WriteString("if rs == nil {\n")
	buf.WriteString("return\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	for _, field := range s.Fields {
		buf.WriteString(generateFieldReset(field, "rs"))
	}

	buf.WriteString("}\n")

	return buf.String()
}

// generateFieldReset генерирует код для сброса одного поля
func generateFieldReset(field fieldInfo, receiver string) string {
	var buf strings.Builder

	fieldRef := fmt.Sprintf("%s.%s", receiver, field.Name)

	// Обрабатываем вложенные структуры с методом Reset()
	if isStructWithReset(field.TypeName, "") {
		// Для указателей на структуры нужна проверка типа на разыменованном значении
		if field.IsPointer {
			buf.WriteString(fmt.Sprintf("if %s != nil {\n", fieldRef))
			buf.WriteString(fmt.Sprintf("if resetter, ok := (*%s).(interface{ Reset() }); ok {\n", fieldRef))
			buf.WriteString(fmt.Sprintf("resetter.Reset()\n"))
			buf.WriteString("}\n")
			buf.WriteString("}\n")
			return buf.String()
		}
		buf.WriteString(fmt.Sprintf("if resetter, ok := %s.(interface{ Reset() }); ok && %s != nil {\n", fieldRef, fieldRef))
		buf.WriteString(fmt.Sprintf("resetter.Reset()\n"))
		buf.WriteString("}\n")
		return buf.String()
	}

	if field.IsPointer {
		buf.WriteString(fmt.Sprintf("if %s != nil {\n", fieldRef))
		buf.WriteString(fmt.Sprintf("*%s", fieldRef))

		// Проверяем, указывает ли указатель на тип с Reset или срез/мапу
		if isPointerTypeWithReset(field.TypeName) {
			buf.WriteString(fmt.Sprintf(".Reset()\n"))
		} else if field.IsSlice {
			buf.WriteString(fmt.Sprintf(" = %s[:0]\n", fieldRef))
		} else if field.IsMap {
			buf.WriteString(fmt.Sprintf("\n"))
			buf.WriteString(fmt.Sprintf("clear(*%s)\n", fieldRef))
		} else {
			buf.WriteString(fmt.Sprintf(" = %s\n", getZeroValue(field.TypeName)))
		}
		buf.WriteString("}\n")
	} else if field.IsSlice {
		buf.WriteString(fmt.Sprintf("%s = %s[:0]\n", fieldRef, fieldRef))
	} else if field.IsMap {
		buf.WriteString(fmt.Sprintf("clear(%s)\n", fieldRef))
	} else {
		buf.WriteString(fmt.Sprintf("%s = %s\n", fieldRef, getZeroValue(field.TypeName)))
	}

	return buf.String()
}

// isStructWithReset проверяет, должен ли тип быть обработан как структура с методом Reset()
func isStructWithReset(typeName string, importPath string) bool {
	// Эти типы не имеют методов Reset() (примитивы, срезы, мапы)
	builtinTypes := map[string]bool{
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"float32":    true,
		"float64":    true,
		"bool":       true,
		"string":     true,
		"error":      true,
		"byte":       true,
		"rune":       true,
		"complex64":  true,
		"complex128": true,
		"slice":      true, // срезы
		"map":        true, // мапы
	}

	return !builtinTypes[typeName] && typeName != ""
}

// isPointerTypeWithReset проверяет, должен ли указатель на тип сбрасываться
func isPointerTypeWithReset(typeName string) bool {
	return typeName != "" && !isBuiltinType(typeName)
}

// isBuiltinType проверяет, является ли тип встроенным типом Go
func isBuiltinType(typeName string) bool {
	builtinTypes := map[string]bool{
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"float32":    true,
		"float64":    true,
		"bool":       true,
		"string":     true,
		"error":      true,
		"byte":       true,
		"rune":       true,
		"complex64":  true,
		"complex128": true,
	}
	return builtinTypes[typeName]
}

// getZeroValue возвращает нулевое значение для заданного типа
func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "0"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "byte":
		return "0"
	case "rune":
		return "0"
	case "complex64", "complex128":
		return "0"
	default:
		return "nil"
	}
}

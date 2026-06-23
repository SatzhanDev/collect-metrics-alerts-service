// Package main реализует генератор методов Reset() для структур Go.
//
// # Использование
//
// Запуск из корня проекта:
//
//	go run ./cmd/reset/... .
//
// Или собрать и запустить:
//
//	go build -o reset ./cmd/reset
//	./reset .
//
// # Как работает
//
// Программа рекурсивно сканирует все пакеты начиная с указанной директории.
// В каждом пакете она ищет структуры, над которыми стоит комментарий:
//
//	// generate:reset
//
// Для каждой такой структуры генерируется метод Reset(), который сбрасывает
// все поля структуры к нулевым значениям. Все сгенерированные методы
// записываются в файл reset.gen.go внутри того же пакета.
//
// # Правила сброса полей
//
//   - Примитивы (int, string, bool и др.) → нулевое значение (0, "", false)
//   - Срезы → срез[:0] (длина обнуляется, память сохраняется)
//   - Мапы → clear(m)
//   - Указатели на примитивы → разыменование и сброс (если не nil)
//   - Указатели на структуры → вызов Reset() через интерфейс (если не nil)
//   - Именованные типы → попытка вызова Reset() через интерфейс
//   - Интерфейсы, каналы, функции → nil
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// zeroValues содержит нулевые значения для примитивных типов Go.
var zeroValues = map[string]string{
	"int": "0", "int8": "0", "int16": "0", "int32": "0", "int64": "0",
	"uint": "0", "uint8": "0", "uint16": "0", "uint32": "0", "uint64": "0",
	"float32": "0", "float64": "0",
	"complex64": "0", "complex128": "0",
	"string": `""`, "bool": "false",
	"byte": "0", "rune": "0", "uintptr": "0",
}

// structField описывает одно поле структуры.
type structField struct {
	Name string
	Type ast.Expr
}

// structInfo описывает структуру, для которой нужно сгенерировать Reset().
type structInfo struct {
	Name     string
	Fields   []structField
	Receiver string
}

// packageInfo описывает пакет со всеми найденными структурами.
type packageInfo struct {
	Dir     string
	Name    string
	Structs []structInfo
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	packages := make(map[string]*packageInfo)

	// Рекурсивно обходим все директории
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем скрытые директории и vendor
		if info.IsDir() {
			name := info.Name()
			if name != "." && (strings.HasPrefix(name, ".") || name == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}

		// Обрабатываем только .go файлы, пропускаем тесты и уже сгенерированные файлы
		if !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		dir := filepath.Dir(path)

		structs, pkgName, parseErr := parseFile(path)
		if parseErr != nil {
			log.Printf("warning: cannot parse %s: %v", path, parseErr)
			return nil
		}
		if len(structs) == 0 {
			return nil
		}

		if _, ok := packages[dir]; !ok {
			packages[dir] = &packageInfo{Dir: dir, Name: pkgName}
		}
		packages[dir].Structs = append(packages[dir].Structs, structs...)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(packages) == 0 {
		fmt.Println("no structs with // generate:reset found")
		return
	}

	for _, pkg := range packages {
		if genErr := generateFile(pkg); genErr != nil {
			log.Fatalf("generate %s: %v", pkg.Dir, genErr)
		}
		fmt.Printf("generated: %s/reset.gen.go\n", pkg.Dir)
	}
}

// parseFile парсит Go-файл и возвращает все структуры с комментарием // generate:reset.
func parseFile(path string) ([]structInfo, string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}

	var structs []structInfo
	pkgName := f.Name.Name

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Проверяем наличие комментария // generate:reset
		if !hasResetComment(genDecl.Doc) {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			info := structInfo{
				Name:     typeSpec.Name.Name,
				Receiver: receiverName(typeSpec.Name.Name),
			}

			// Собираем поля структуры
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					info.Fields = append(info.Fields, structField{
						Name: name.Name,
						Type: field.Type,
					})
				}
			}

			structs = append(structs, info)
		}
	}

	return structs, pkgName, nil
}

// hasResetComment проверяет, содержит ли группа комментариев // generate:reset.
func hasResetComment(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if c.Text == "// generate:reset" {
			return true
		}
	}
	return false
}

// receiverName возвращает имя получателя метода — первая строчная буква имени типа.
func receiverName(typeName string) string {
	if typeName == "" {
		return "s"
	}
	return strings.ToLower(string(typeName[0]))
}

// generateFile создаёт файл reset.gen.go в директории пакета.
func generateFile(pkg *packageInfo) error {
	var buf bytes.Buffer

	buf.WriteString("// Code generated by reset generator. DO NOT EDIT.\n\n")
	buf.WriteString("package " + pkg.Name + "\n\n")

	for _, s := range pkg.Structs {
		generateMethod(&buf, s)
	}

	// Форматируем сгенерированный код через go/format
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Записываем неотформатированный код для отладки
		outPath := filepath.Join(pkg.Dir, "reset.gen.go")
		_ = os.WriteFile(outPath, buf.Bytes(), 0644)
		return fmt.Errorf("format source: %w", err)
	}

	outPath := filepath.Join(pkg.Dir, "reset.gen.go")
	return os.WriteFile(outPath, formatted, 0644)
}

// generateMethod записывает в buf метод Reset() для одной структуры.
func generateMethod(buf *bytes.Buffer, s structInfo) {
	recv := s.Receiver
	fmt.Fprintf(buf, "func (%s *%s) Reset() {\n", recv, s.Name)
	fmt.Fprintf(buf, "\tif %s == nil {\n\t\treturn\n\t}\n", recv)

	if len(s.Fields) > 0 {
		buf.WriteString("\n")
		for _, field := range s.Fields {
			line := fieldResetLine(recv+"."+field.Name, field.Type)
			buf.WriteString(line + "\n")
		}
	}

	buf.WriteString("}\n\n")
}

// fieldResetLine генерирует строку кода для сброса одного поля структуры.
//
// Правила:
//   - примитив → нулевое значение
//   - срез → [:0]
//   - мапа → clear()
//   - указатель на примитив → проверка nil + разыменование
//   - указатель на тип → проверка nil + вызов Reset() через интерфейс
//   - именованный тип → вызов Reset() через интерфейс (с взятием адреса)
//   - интерфейс/канал/функция → nil
func fieldResetLine(qualified string, typ ast.Expr) string {
	switch t := typ.(type) {

	case *ast.Ident:
		// Примитивный тип
		if zero, ok := zeroValues[t.Name]; ok {
			return "\t" + qualified + " = " + zero
		}
		// Именованный тип (структура) — вызываем Reset() если есть
		return fmt.Sprintf(
			"\tif resetter, ok := any(&%s).(interface{ Reset() }); ok {\n\t\tresetter.Reset()\n\t}",
			qualified,
		)

	case *ast.StarExpr:
		// Указатель на примитив
		if inner, ok := t.X.(*ast.Ident); ok {
			if zero, ok2 := zeroValues[inner.Name]; ok2 {
				return fmt.Sprintf(
					"\tif %s != nil {\n\t\t*%s = %s\n\t}",
					qualified, qualified, zero,
				)
			}
		}
		// Указатель на именованный/составной тип
		return fmt.Sprintf(
			"\tif resetter, ok := any(%s).(interface{ Reset() }); ok && %s != nil {\n\t\tresetter.Reset()\n\t}",
			qualified, qualified,
		)

	case *ast.ArrayType:
		if t.Len == nil {
			// Срез: обрезаем до нулевой длины, память сохраняем
			return "\t" + qualified + " = " + qualified + "[:0]"
		}
		// Массив фиксированной длины — не в спецификации, пропускаем
		return "\t// array field skipped: " + qualified

	case *ast.MapType:
		return "\tclear(" + qualified + ")"

	case *ast.SelectorExpr:
		// Тип из другого пакета (например, time.Time, sync.Mutex)
		return fmt.Sprintf(
			"\tif resetter, ok := any(&%s).(interface{ Reset() }); ok {\n\t\tresetter.Reset()\n\t}",
			qualified,
		)

	case *ast.InterfaceType:
		return "\t" + qualified + " = nil"

	case *ast.ChanType:
		return "\t" + qualified + " = nil"

	case *ast.FuncType:
		return "\t" + qualified + " = nil"

	default:
		return "\t// TODO: unsupported type for field " + qualified
	}
}

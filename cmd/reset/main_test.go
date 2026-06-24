package main

import (
	"bytes"
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── hasResetComment ───────────────────────────────────────────────────────

func TestHasResetComment_Nil(t *testing.T) {
	assert.False(t, hasResetComment(nil))
}

func TestHasResetComment_WithResetComment(t *testing.T) {
	cg := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// generate:reset"},
		},
	}
	assert.True(t, hasResetComment(cg))
}

func TestHasResetComment_OtherComment(t *testing.T) {
	cg := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// some unrelated comment"},
		},
	}
	assert.False(t, hasResetComment(cg))
}

func TestHasResetComment_MultipleComments(t *testing.T) {
	cg := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// first"},
			{Text: "// generate:reset"},
			{Text: "// last"},
		},
	}
	assert.True(t, hasResetComment(cg))
}

// ─── receiverName ─────────────────────────────────────────────────────────

func TestReceiverName_Normal(t *testing.T) {
	assert.Equal(t, "m", receiverName("Metrics"))
	assert.Equal(t, "e", receiverName("Event"))
	assert.Equal(t, "s", receiverName("Server"))
}

func TestReceiverName_Empty(t *testing.T) {
	assert.Equal(t, "s", receiverName(""))
}

func TestReceiverName_Lowercase(t *testing.T) {
	assert.Equal(t, "m", receiverName("myStruct"))
}

// ─── fieldResetLine ────────────────────────────────────────────────────────

func TestFieldResetLine_PrimitiveInt(t *testing.T) {
	result := fieldResetLine("s.Count", &ast.Ident{Name: "int64"})
	assert.Equal(t, "\ts.Count = 0", result)
}

func TestFieldResetLine_PrimitiveBool(t *testing.T) {
	result := fieldResetLine("s.Flag", &ast.Ident{Name: "bool"})
	assert.Equal(t, "\ts.Flag = false", result)
}

func TestFieldResetLine_PrimitiveString(t *testing.T) {
	result := fieldResetLine("s.Name", &ast.Ident{Name: "string"})
	assert.Equal(t, "\ts.Name = \"\"", result)
}

func TestFieldResetLine_NamedType(t *testing.T) {
	// Именованный тип (структура) → вызов Reset() через интерфейс
	result := fieldResetLine("s.Sub", &ast.Ident{Name: "SubStruct"})
	assert.Contains(t, result, "Reset()")
	assert.Contains(t, result, "resetter")
}

func TestFieldResetLine_PointerToPrimitive(t *testing.T) {
	result := fieldResetLine("s.P", &ast.StarExpr{
		X: &ast.Ident{Name: "int64"},
	})
	assert.Contains(t, result, "*s.P = 0")
	assert.Contains(t, result, "s.P != nil")
}

func TestFieldResetLine_PointerToNamedType(t *testing.T) {
	result := fieldResetLine("s.P", &ast.StarExpr{
		X: &ast.Ident{Name: "MyType"},
	})
	assert.Contains(t, result, "Reset()")
	assert.Contains(t, result, "s.P != nil")
}

func TestFieldResetLine_PointerToSelectorExpr(t *testing.T) {
	// Указатель на тип из другого пакета
	result := fieldResetLine("s.P", &ast.StarExpr{
		X: &ast.SelectorExpr{},
	})
	assert.Contains(t, result, "Reset()")
}

func TestFieldResetLine_Slice(t *testing.T) {
	result := fieldResetLine("s.Items", &ast.ArrayType{
		Len: nil,
		Elt: &ast.Ident{Name: "string"},
	})
	assert.Equal(t, "\ts.Items = s.Items[:0]", result)
}

func TestFieldResetLine_FixedArray(t *testing.T) {
	result := fieldResetLine("s.Arr", &ast.ArrayType{
		Len: &ast.BasicLit{},
		Elt: &ast.Ident{Name: "byte"},
	})
	assert.Contains(t, result, "skipped")
}

func TestFieldResetLine_Map(t *testing.T) {
	result := fieldResetLine("s.M", &ast.MapType{})
	assert.Equal(t, "\tclear(s.M)", result)
}

func TestFieldResetLine_SelectorExpr(t *testing.T) {
	// Тип из другого пакета (time.Time, sync.Mutex и т.д.)
	result := fieldResetLine("s.T", &ast.SelectorExpr{})
	assert.Contains(t, result, "Reset()")
	assert.Contains(t, result, "resetter")
}

func TestFieldResetLine_Interface(t *testing.T) {
	result := fieldResetLine("s.I", &ast.InterfaceType{
		Methods: &ast.FieldList{},
	})
	assert.Equal(t, "\ts.I = nil", result)
}

func TestFieldResetLine_Chan(t *testing.T) {
	result := fieldResetLine("s.C", &ast.ChanType{})
	assert.Equal(t, "\ts.C = nil", result)
}

func TestFieldResetLine_Func(t *testing.T) {
	result := fieldResetLine("s.F", &ast.FuncType{})
	assert.Equal(t, "\ts.F = nil", result)
}

func TestFieldResetLine_Default_UnsupportedType(t *testing.T) {
	// BadExpr не входит ни в один из case → default
	result := fieldResetLine("s.X", &ast.BadExpr{})
	assert.Contains(t, result, "TODO")
}

// ─── generateMethod ────────────────────────────────────────────────────────

func TestGenerateMethod_WithFields(t *testing.T) {
	var buf bytes.Buffer
	s := structInfo{
		Name:     "Event",
		Receiver: "e",
		Fields: []structField{
			{Name: "Count", Type: &ast.Ident{Name: "int64"}},
			{Name: "Name", Type: &ast.Ident{Name: "string"}},
			{Name: "Tags", Type: &ast.ArrayType{Len: nil, Elt: &ast.Ident{Name: "string"}}},
		},
	}
	generateMethod(&buf, s)

	result := buf.String()
	assert.Contains(t, result, "func (e *Event) Reset()")
	assert.Contains(t, result, "e.Count = 0")
	assert.Contains(t, result, `e.Name = ""`)
	assert.Contains(t, result, "e.Tags = e.Tags[:0]")
	assert.Contains(t, result, "if e == nil")
}

func TestGenerateMethod_NoFields(t *testing.T) {
	var buf bytes.Buffer
	s := structInfo{
		Name:     "Empty",
		Receiver: "e",
		Fields:   nil,
	}
	generateMethod(&buf, s)

	result := buf.String()
	assert.Contains(t, result, "func (e *Empty) Reset()")
	assert.Contains(t, result, "if e == nil")
}

// ─── parseFile ────────────────────────────────────────────────────────────

func TestParseFile_WithResetComment(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	content := `package testpkg

// generate:reset
type MyStruct struct {
	Count int64
	Name  string
}
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	structs, pkgName, err := parseFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "testpkg", pkgName)
	require.Len(t, structs, 1)
	assert.Equal(t, "MyStruct", structs[0].Name)
	assert.Equal(t, "m", structs[0].Receiver)
	assert.Len(t, structs[0].Fields, 2)
}

func TestParseFile_NoResetComment(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	content := `package testpkg

type MyStruct struct {
	Count int64
}
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	structs, pkgName, err := parseFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "testpkg", pkgName)
	assert.Empty(t, structs)
}

func TestParseFile_InvalidGo(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.go")
	require.NoError(t, os.WriteFile(tmpFile, []byte("not valid go code {{{"), 0644))

	_, _, err := parseFile(tmpFile)
	require.Error(t, err)
}

func TestParseFile_NonStructType(t *testing.T) {
	// type alias с комментарием — не структура, должна быть пропущена
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	content := `package testpkg

// generate:reset
type MyAlias int
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	structs, _, err := parseFile(tmpFile)
	require.NoError(t, err)
	assert.Empty(t, structs)
}

func TestParseFile_MultipleStructs(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	content := `package testpkg

// generate:reset
type First struct {
	X int
}

// generate:reset
type Second struct {
	Y string
}

type Third struct {
	Z bool
}
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	structs, _, err := parseFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, structs, 2)
	assert.Equal(t, "First", structs[0].Name)
	assert.Equal(t, "Second", structs[1].Name)
}

// ─── generateFile ─────────────────────────────────────────────────────────

func TestGenerateFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	pkg := &packageInfo{
		Dir:  tmpDir,
		Name: "testpkg",
		Structs: []structInfo{
			{
				Name:     "MyType",
				Receiver: "m",
				Fields: []structField{
					{Name: "Value", Type: &ast.Ident{Name: "int64"}},
					{Name: "Items", Type: &ast.ArrayType{Len: nil, Elt: &ast.Ident{Name: "string"}}},
					{Name: "Meta", Type: &ast.MapType{
						Key:   &ast.Ident{Name: "string"},
						Value: &ast.Ident{Name: "string"},
					}},
				},
			},
		},
	}

	err := generateFile(pkg)
	require.NoError(t, err)

	outPath := filepath.Join(tmpDir, "reset.gen.go")
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "package testpkg")
	assert.Contains(t, content, "func (m *MyType) Reset()")
	assert.Contains(t, content, "m.Value = 0")
	assert.Contains(t, content, "m.Items = m.Items[:0]")
	assert.Contains(t, content, "clear(m.Meta)")
}

func TestGenerateFile_MultipleStructs(t *testing.T) {
	tmpDir := t.TempDir()
	pkg := &packageInfo{
		Dir:  tmpDir,
		Name: "multipkg",
		Structs: []structInfo{
			{
				Name:     "Alpha",
				Receiver: "a",
				Fields:   []structField{{Name: "N", Type: &ast.Ident{Name: "int"}}},
			},
			{
				Name:     "Beta",
				Receiver: "b",
				Fields:   []structField{{Name: "S", Type: &ast.Ident{Name: "string"}}},
			},
		},
	}

	require.NoError(t, generateFile(pkg))

	data, err := os.ReadFile(filepath.Join(tmpDir, "reset.gen.go"))
	require.NoError(t, err)

	s := string(data)
	assert.Contains(t, s, "func (a *Alpha) Reset()")
	assert.Contains(t, s, "func (b *Beta) Reset()")
}

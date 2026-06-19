package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer запрещает прямой вызов os.Exit в функции main пакета main.
//
// Прямой вызов os.Exit обходит все отложенные вызовы (defer), что может
// привести к утечкам ресурсов, незакрытым соединениям и некорректному
// завершению программы. Вместо os.Exit рекомендуется использовать
// graceful shutdown или возвращать ошибки через канал/переменную.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			// Ищем функцию main
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			// Обходим тело функции main в поисках вызовов os.Exit
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Проверяем, что вызов имеет вид selExpr.X.selExpr.Sel — т.е. пакет.Функция
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Pos(), "прямой вызов os.Exit запрещён в функции main пакета main")
				}

				return true
			})
		}
	}

	return nil, nil
}

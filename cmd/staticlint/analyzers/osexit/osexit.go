// Package osexit предоставляет анализатор, запрещающий прямые вызовы os.Exit в функции main.
//
// Этот анализатор является частью инструмента staticlint.
//
// Анализатор проверяет, что в функции main() пакета main нет прямых вызовов os.Exit().
// Это best practice по следующим причинам:
//   - Прямые вызовы os.Exit() обходят отложенные функции (defer) и код очистки
//   - Затрудняется тестирование, так как программа завершается внезапно
//   - Вместо этого следует использовать стандартные паттерны обработки ошибок Go
//
// Если os.Exit необходим, рассмотрите:
//   - Использование log.Fatal() или log.Fatal().Err() для фатальных ошибок (они внутренне вызывают os.Exit)
//   - Возврат ошибок из main() и их обработку на более высоком уровне
//   - Использование defer для очистки перед любым выходом
//
// Анализатор применяется только к пакету main и проверяет только функцию main().
package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer — пользовательский анализатор, обнаруживающий прямые вызовы os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "обнаруживает прямые вызовы os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Пропускаем файлы из кэша Go (сгенерированные тестовые файлы вида .../Library/Caches/go-build/*-d)
	for _, file := range pass.Files {
		if pass.Fset == nil {
			continue
		}
		pos := pass.Fset.Position(file.Pos())
		// Пропускаем кэш-файлы (имеют суффикс -d и путь go-build)
		if len(pos.Filename) > 2 && pos.Filename[len(pos.Filename)-2:] == "-d" {
			return nil, nil
		}
		if len(pos.Filename) > len("/Library/Caches/go-build/") && pos.Filename[len(pos.Filename)-2:] == "-d" {
			return nil, nil
		}
		// Также проверяем префикс пути к кэшу
		if len(pos.Filename) > len("/Library/Caches/go-build") && pos.Filename[:len("/Library/Caches/go-build")] == "/Library/Caches/go-build" {
			return nil, nil
		}
	}

	// Сначала находим функцию main
	var mainFunc *ast.FuncDecl
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				mainFunc = fn
				break
			}
		}
	}

	// Если функция main не найдена, нечего проверять
	if mainFunc == nil {
		return nil, nil
	}

	// Обходим AST тела функции main в поисках вызовов os.Exit
	ast.Inspect(mainFunc.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Проверяем, является ли это вызовом os.Exit
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && sel.Sel.Name == "Exit" {
			pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещён")
		}

		return true
	})

	return nil, nil
}

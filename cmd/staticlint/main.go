// Package main реализует мультичекер, объединяющий несколько инструментов статического анализа.
//
// Использование:
//
//	staticlint [флаги] [пакеты]
//
// Мультичекер запускает следующие инструменты анализа:
//
// Стандартные анализы из golang.org/x/tools/go/analysis/passes:
//   - asmdecl: обнаруживает несоответствия языка ассемблера с объявлениями Go
//   - assign: обнаруживает невозможные присваивания
//   - atomic: обнаруживает распространённые ошибки в работе с атомарными операциями
//   - atomicalign: обнаруживает проблемы с выравниванием атомарных операций
//   - bools: обнаруживает ошибки в булевых операторах
//   - buildtag: обнаруживает некорректное использование build-тегов
//   - cgocall: обнаруживает проблемное использование cgo
//   - copylock: обнаруживает случайное использование заблокированных объектов
//   - directive: обнаруживает проблемы с директивами компилятора
//   - httpresponse: обнаруживает ошибки в работе с http.ResponseWriter
//   - ifaceassert: обнаруживает невозможные приведения типов интерфейсов
//   - nilness: обнаруживает недостижимые проверки на nil
//   - pkgfact: обнаруживает проблемы с факторизацией пакетов
//   - printf: обнаруживает ошибки в вызовах printf
//   - shadow: обнаруживает перекрытие переменных
//   - shift: обнаруживает невозможные сдвиги
//   - stdmethods: обнаруживает ошибки в стандартных методах
//   - stringintconv: обнаруживает проблемы с преобразованиями string/int
//   - structtag: обнаруживает ошибки в тегах структур
//   - tests: обнаруживает проблемы в тестах
//   - unsafeptr: обнаруживает невозможные преобразования указателей
//   - unusedresult: обнаруживает неиспользуемые результаты вызовов функций
//
// Анализаторы класса SA (staticcheck.io):
//   - SA1000: некорректное использование unsafe
//   - SA2000: некорректное использование testing
//   - SA3000: некорректное использование math/big
//   - SA4000-SF6095: различные анализы для криптографии, сетевых пакетов, JSON/XML, регулярных выражений и других
//
// Сторонние анализаторы:
//   - durationcheck: проверка на ошибки в работе с временными интервалами
//   - errcheck: проверка на необработанные ошибки
//   - simple: простые проблемы в коде
//   - staticcheck: общий статический анализ от staticcheck.io
//   - unused: неиспользуемые переменные и импорты
//
// Пользовательские анализаторы:
//   - osexit: обнаруживает прямые вызовы os.Exit в функции main пакета main
//
// Примеры:
//
//	Запустить все проверки на текущем пакете:
//	  staticlint ./...
//
//	Запустить проверки на конкретных пакетах:
//	  staticlint ./cmd/server ./internal/handler
//
//	Запустить проверки с подробным выводом:
//	  staticlint -v ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	// Стандартные анализы из golang.org/x/tools/go/analysis/passes
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	// Пользовательский анализатор osexit
	osexit "github.com/KaziPHone/go-musthave-metrics-tpl/cmd/staticlint/analyzers/osexit"
)

// analyzers — список всех анализаторов для запуска
var analyzers = []*analysis.Analyzer{
	// Стандартные анализы из golang.org/x/tools/go/analysis/passes
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	atomicalign.Analyzer,
	bools.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	copylock.Analyzer,
	directive.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	nilness.Analyzer,
	pkgfact.Analyzer,
	printf.Analyzer,
	shadow.Analyzer,
	shift.Analyzer,
	stdmethods.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	tests.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,

	// Пользовательский анализатор osexit
	osexit.Analyzer,
}

func main() {
	multichecker.Main(analyzers...)
}

/*
Package main реализует multichecker — инструмент статического анализа кода
для проекта collect-metrics-alerts-service.

# Запуск

Собрать бинарник и запустить:

	go build -o staticlint ./cmd/staticlint
	./staticlint ./...

Или запустить напрямую через go run:

	go run ./cmd/staticlint/... ./...

Для получения справки по флагам:

	./staticlint -help

Для объяснения конкретного анализатора:

	./staticlint -explain SA1000

# Анализаторы

## Стандартные анализаторы (golang.org/x/tools/go/analysis/passes)

  - asmdecl: проверяет соответствие ассемблерных деклараций Go-функциям
  - assign: обнаруживает бесполезные присваивания (x = x)
  - atomic: проверяет корректное использование sync/atomic
  - bools: находит ошибки с булевыми операторами (&&, ||)
  - buildtag: проверяет корректность тегов сборки //go:build
  - cgocall: обнаруживает нарушения правил передачи указателей cgo
  - copylock: обнаруживает копирование типов, содержащих мьютексы
  - errorsas: проверяет второй аргумент errors.As (должен быть указателем)
  - httpresponse: проверяет корректное использование HTTP-ответов
  - ifaceassert: обнаруживает невозможные type assertion для интерфейсов
  - loopclosure: проверяет использование переменных цикла в замыканиях горутин
  - lostcancel: проверяет, что cancel-функция контекста всегда вызывается
  - nilfunc: обнаруживает сравнение функций с nil (всегда false)
  - printf: проверяет форматные строки Printf-подобных функций
  - shadow: обнаруживает затенение переменных во вложенных блоках
  - shift: проверяет побитовые сдвиги (не выходят ли за ширину типа)
  - stdmethods: проверяет сигнатуры методов стандартных интерфейсов (Error, String и др.)
  - stringintconv: обнаруживает подозрительное преобразование int → string
  - structtag: проверяет синтаксис тегов полей структур
  - tests: проверяет корректность сигнатур тестовых функций и примеров
  - unmarshal: проверяет передачу non-pointer в Unmarshal-функции
  - unreachable: обнаруживает недостижимый код после return/panic/break
  - unsafeptr: обнаруживает некорректные преобразования uintptr в unsafe.Pointer
  - unusedresult: проверяет, что результаты некоторых функций не игнорируются

## Анализаторы SA (honnef.co/go/tools/staticcheck)

Все анализаторы класса SA из пакета staticcheck:
  - SA1xxx: неправильное использование стандартных библиотек
  - SA2xxx: проблемы с многопоточностью (горутины, мьютексы, каналы)
  - SA3xxx: проблемы с тестами
  - SA4xxx: бесполезный код (мёртвый код, всегда-true/false условия)
  - SA5xxx: ошибочный код (nil dereference, неверные форматные строки)
  - SA6xxx: проблемы с производительностью
  - SA9xxx: сомнительные конструкции с высокой вероятностью ошибки

## Анализаторы S1 (honnef.co/go/tools/simple)

Анализаторы класса S1 предлагают упрощения кода:
замена сложных конструкций на более идиоматичные Go-аналоги.
Например: использование strings.Contains вместо strings.Index >= 0.

## Анализаторы ST1 (honnef.co/go/tools/stylecheck)

Анализаторы класса ST1 проверяют стилистические соглашения:
именование (ST1003), стиль комментариев, использование
http.StatusXxx констант вместо числовых кодов (ST1013) и др.

## Публичные анализаторы

  - ineffassign (github.com/gordonklaus/ineffassign): обнаруживает присваивания
    переменным, значения которых никогда не используются после присваивания.
    Помогает найти опечатки и логические ошибки.

  - bodyclose (github.com/timakin/bodyclose): проверяет, что тело HTTP-ответа
    (resp.Body) всегда закрывается. Незакрытое тело приводит к утечке
    горутин и соединений.

## Собственный анализатор

  - noosexit: запрещает прямой вызов os.Exit в функции main пакета main.
    Прямой вызов os.Exit обходит все отложенные вызовы (defer), что может
    привести к утечкам ресурсов и некорректному завершению программы.
    Рекомендуется использовать graceful shutdown через каналы сигналов.
*/
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	analyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		// Публичные анализаторы
		ineffassign.Analyzer,
		bodyclose.Analyzer,

		// Собственный анализатор
		NoOsExitAnalyzer,
	}

	// Все SA-анализаторы из staticcheck.io
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// S1-анализаторы (упрощения кода)
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// ST1-анализаторы (стиль кода)
	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	multichecker.Main(analyzers...)
}

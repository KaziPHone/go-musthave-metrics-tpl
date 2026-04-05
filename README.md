# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Бенчмарки и профилирование

В проекте реализованы бенчмарки для измерения производительности ключевых компонентов:

### Бенчмарки (запуск: `go test -bench=. -benchmem`)

| Компонент | Операции/с | Байт/операцию | Аллокаций/операцию |
|-----------|------------|---------------|-------------------|
| UpdateValueHandler (Gauge) | 3,991,144 | 56 B/op | 2 allocs/op |
| UpdateValueHandler (Counter) | 4,080,225 | 56 B/op | 2 allocs/op |
| UpdateHandler | 1,326,811 | 1,094 B/op | 14 allocs/op |
| UpdatesHandler | 176,292 | 3,752 B/op | 52 allocs/op |
| GetMetricHandler | 9,645,700 | 0 B/op | 0 allocs/op |
| ListMetricsHandler | 68,125 | 10,288 B/op | 200 allocs/op |
| CalcSHA256Hash | 2,076,225 | 120 B/op | 3 allocs/op |
| CalcSHA256HashBuffer | 1,982,108 | 168 B/op | 4 allocs/op |
| Storage.UpdateMetric | 3,598,002 | 142 B/op | 4 allocs/op |
| Storage.GetMetric | 18,262,268 | 21 B/op | 1 alloc/op |

### Профилирование памяти

**Процедура:**
1. Снимите базовый профиль: `go run profile/profile_runner.go`
2. Снимите профиль после нагрузки: `go test -bench=. -benchmem -memprofile=profiles/result.memprof`
3. Сравните профили: `go tool pprof -base profiles/base.memprof -top profiles/result.memprof`

**Результаты оптимизации:**

После оптимизации использования памяти:

```
Showing nodes accounting for -1605.56MB, 1.48% of 108613.92MB total
...
-1920.26MB  1.77%  compress/flate.(*dictDecoder).init (inline)
-430.17MB   0.4%  compress/flate.NewReader
-409.62MB  0.38%  github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage.(*MStorage).updateMetricMemory
-406.55MB  0.37%  crypto/internal/fips140/sha256.New (inline)
-154.65MB  0.14%  io.ReadAll
...
```

Отрицательные значения показывают, что память уменьшилась после оптимизации.

**Оптимизации:**

1. **Pool для хеширования** (`pkg/helpers/shaHelper.go`): Использование `sync.Pool` для `sha256.Hash` снижает аллокации для повторных вызовов.

2. **Pool для GetMetrics** (`pkg/helpers/auditHelper.go`): Повторное использование буфера среза для сбора метрик уменьшает аллокации.

3. **Оптимизация GetClientIP** (`pkg/helpers/auditHelper.go`): Замена `strings.Split` и `strings.TrimSpace` на прямую работу с строками снижает аллокации.

4. **Уменьшение проверок nil** (`pkg/storage/memoryStorage.go`): Упрощение проверок pointer-значений уменьшает количество операций.

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

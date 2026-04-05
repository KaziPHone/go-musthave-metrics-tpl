# Reset Tool

Генератор методов `Reset()` для структур Go.

## Установка

```bash
go build -o reset ./cmd/reset
```

## Использование

```bash
reset <root_directory>
```

Где `<root_directory>` — корневая директория проекта.

## Генерация метода Reset()

Чтобы сгенерировать метод `Reset()` для структуры, добавьте комментарий `// generate:reset` перед её объявлением:

```go
// generate:reset
type MyStruct struct {
    Field1 int
    Field2 string
}
```

## Правила генерации

- **Примитивы**: сбрасываются к нулевым значениям (`0`, `""`, `false`)
- **Срезы**: обрезаются до нулевой длины (`s = s[:0]`)
- **Мапы**: очищаются (`clear(m)`)
- **Указатели**: если не `nil`, сбрасываются по правилам для типа
- **Вложенные структуры**: вызывают свой метод `Reset()` если он существует

## Пример

```go
// generate:reset
type ResetableStruct struct {
    i     int
    str   string
    strP  *string
    s     []int
    m     map[string]string
    child *ResetableStruct
}
```

Генерирует:

```go
func (rs *ResetableStruct) Reset() {
    if rs == nil {
        return
    }

    rs.i = 0
    rs.str = ""
    if rs.strP != nil {
        *rs.strP = ""
    }
    rs.s = rs.s[:0]
    clear(rs.m)
    if rs.child != nil {
        if resetter, ok := (*rs.child).(interface{ Reset() }); ok {
            resetter.Reset()
        }
    }
}
```

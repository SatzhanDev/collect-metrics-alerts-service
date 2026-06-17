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

## Профилирование (pprof)

Профили сняты под нагрузкой бенчмарком `BenchmarkMemStorage_UpdateBatch` (1000 метрик за итерацию).
Файлы хранятся в папке `profiles/`.

### Как воспроизвести

```bash
# Снять базовый профиль CPU
go test ./internal/repository/mem/... -bench=BenchmarkMemStorage_UpdateBatch -benchtime=30s -cpuprofile=profiles/base.pprof

# Снять базовый профиль аллокаций
go test ./internal/repository/mem/... -bench=BenchmarkMemStorage_UpdateBatch -benchtime=30s -memprofile=profiles/base_allocs.pprof

# После оптимизации снять итоговый профиль CPU
go test ./internal/repository/mem/... -bench=BenchmarkMemStorage_UpdateBatch -benchtime=30s -cpuprofile=profiles/result.pprof

# Сравнить профили
go tool pprof -diff_base=profiles/base.pprof profiles/result.pprof
```

### Результаты оптимизации

Узкое место обнаружено в `MemStorage.UpdateBatch`: при обработке каждой метрики выполнялась блокировка мьютекса отдельно для каждой записи, что создавало высокий contention под нагрузкой.

**Оптимизация:** переход к единой блокировке на весь batch вместо блокировки per-record.

Вывод `go tool pprof -diff_base`:

```
(pprof) top10
Showing nodes accounting for -1.2s, 8.57% of 14s total
      flat  flat%   sum%        cum   cum%
    -0.60s  4.29%  4.29%     -0.60s  4.29%  sync.(*Mutex).Lock
    -0.60s  4.29%  8.57%     -0.60s  4.29%  sync.(*Mutex).Unlock
```

CPU-время на операциях блокировки сократилось на ~8.5%.

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

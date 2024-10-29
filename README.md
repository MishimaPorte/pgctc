# Генератор парсеров для композитных типов постгреса

## Что куда как + референсы
- [кастомные типы](https://www.postgresql.org/docs/current/rowtypes.html)
- [релевантная ишью в алхимии](https://github.com/sqlalchemy/sqlalchemy/discussions/10101)
- [sqlalchemy_utils](https://github.com/kvesteri/sqlalchemy-utils)

В голанге предполагается либо парсить постгревый аутпут руками, либо руками генерировать парсер.

## Как пользоваться

В папке `example` пример скрипта (`metagenerator.go`), который имплементит интерфейсы `driver.Valuer` и `driver.Scanner` для типов и для всех водящих в эти типы типов.

БЕЗ зависимостей бтв.

package shell

import "fmt"

// makeDirectory проверяет единственный обязательный путь и создаёт каталог, возвращая пустой вывод.
func (sh *Shell) makeDirectory(args []string) (Result, error) {
	const op = "shell.mkdir"
	name, err := singlePath(args, "")
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	if name == "" {
		return Result{}, fmt.Errorf("%s: требуется путь нового каталога", op)
	}
	if err := sh.filesystem.MakeDirectory(sh.current, name); err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	return Result{}, nil
}

// changeMode проверяет режим и путь, затем меняет права узла, возвращая пустой вывод либо ошибку.
func (sh *Shell) changeMode(args []string) (Result, error) {
	const op = "shell.chmod"
	const argumentCount = 2
	if len(args) != argumentCount {
		return Result{}, fmt.Errorf("%s: ожидаются права и один путь", op)
	}
	name, err := singlePath(args[1:], "")
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	if err := sh.filesystem.ChangeMode(sh.current, name, args[0]); err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	return Result{}, nil
}

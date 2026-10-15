package shell

import "fmt"

// changeDirectory меняет текущий каталог только после успешного обхода всего пути; возвращает пустой вывод.
func (sh *Shell) changeDirectory(args []string) (Result, error) {
	const op = "shell.cd"
	path, err := singlePath(args, "/")
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	stack, err := sh.filesystem.Resolve(sh.current, path)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	if !stack[len(stack)-1].Mode.IsDir() {
		return Result{}, fmt.Errorf("%s: %q: не каталог", op, path)
	}
	sh.current = stack
	return Result{}, nil
}

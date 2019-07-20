package metrics

import (
	"fmt"
	"os"
	"runtime"
)

func recoverPanic() {
	var recoverResult interface{}
	if recoverResult = recover(); recoverResult == nil {
		return
	}

	var err error
	switch recoverResult := recoverResult.(type) {
	case error:
		err = recoverResult
	default:
		err = fmt.Errorf("%v", recoverResult)
	}

	buf := make([]byte, 1<<16)
	stack := runtime.Stack(buf, true)

	_, _ = fmt.Fprintf(os.Stderr, "[panic] %s\n%s\n", err, string(stack))
}

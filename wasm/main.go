package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/node"
)

func parse(filename string, contents string) (rootNode node.Node, w *linter.RootWalker, err error) {
	rootNode, w, err = linter.ParseContents(filename, []byte(contents), "UTF-8", nil)
	if err != nil {
		return nil, nil, err
	}

	if !meta.IsIndexingComplete() {
		w.UpdateMetaInfo()
	}

	return rootNode, w, nil
}

func getReports(contents string) ([]vscode.Diagnostic, error) {
	meta.ResetInfo()
	if _, _, err := parse(`demo.php`, contents); err != nil {
		return nil, err
	}
	meta.SetIndexingComplete(true)
	_, w, err := parse(`demo.php`, contents)
	if err != nil {
		return nil, err
	}
	return w.Diagnostics, err
}

func main() {
	linter.LangServer = true

	go linter.MemoryLimiterThread()

	js.Global().Set("analyzeCallback", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		text := js.Global().Get("editor").Call("getValue").String()
		diags, err := getReports(text)

		var value string
		if err != nil {
			value = "ERROR: " + err.Error()
		} else {
			m, _ := json.Marshal(diags)
			value = string(m)
		}

		js.Global().Call("showErrors", value)
		return nil
	}))

	select {}
}

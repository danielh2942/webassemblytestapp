//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

// prettyJson makes any json string inputted pretty.
func prettyJson(input string) (string, error) {
	var raw any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "    ")
	if err != nil {
		return "", err
	}

	return string(pretty), nil
}

// jsonWrapper returns a js wrapper function to allow interop between go and js
func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return map[string]any {
				"error":"Invalid number of arguments passed",
			}
		}
		// Get access to the DOM
		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			return map[string]any {
				"error":"Unable to get document object",
			}
		}
		// Get Access to the output text area.
		jsonOutputTextArea := jsDoc.Call("getElementById","jsonoutput")
		if !jsonOutputTextArea.Truthy() {
			return map[string]any {
				"error":"Unable to get output object",
			}
		}

		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		pretty, err := prettyJson(inputJSON)
		if err != nil {
			errStr := fmt.Sprintf("Unable to convert json: %s", err)
			return map[string]any {
				"error":errStr,
			}
		}
		// Alter the DOM from within the wasm itself
		jsonOutputTextArea.Set("value",pretty)
		return nil
	})
	return jsonFunc
}

func main() {
	fmt.Println("Hello World!")
	// Make the function reachable in the JS code
	js.Global().Set("formatJSON", jsonWrapper())
	// Make the main program not terminate so formatJSON can be called at all.
	select {}
}


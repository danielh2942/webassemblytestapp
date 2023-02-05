//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			return map[string]any{
				"error": "Invalid number of arguments passed",
			}
		}
		// Get access to the DOM
		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			return map[string]any{
				"error": "Unable to get document object",
			}
		}
		// Get Access to the output text area.
		jsonOutputTextArea := jsDoc.Call("getElementById", "jsonoutput")
		if !jsonOutputTextArea.Truthy() {
			return map[string]any{
				"error": "Unable to get output object",
			}
		}

		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		pretty, err := prettyJson(inputJSON)
		if err != nil {
			errStr := fmt.Sprintf("Unable to convert json: %s", err)
			return map[string]any{
				"error": errStr,
			}
		}
		// Alter the DOM from within the wasm itself
		jsonOutputTextArea.Set("value", pretty)
		return nil
	})
	return jsonFunc
}

// check sending HTTP Requests from inside wasm
func ping() js.Func {
	pingFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		// as this is asynchronous it requires a promise and as a result uses resolve/reject syntax
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			reject := args[1]
			// make http request non blocking
			go func() {
				res, err := http.DefaultClient.Get("ping")
				// Check if the server can handle the ping request
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				defer res.Body.Close()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				resolve.Invoke(string(data))
			}()
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
	return pingFunc
}

func main() {
	fmt.Println("Hello World!")
	// Make the function reachable in the JS code
	js.Global().Set("formatJSON", jsonWrapper())
	// Make ping function visible to anyone who can access the server.
	js.Global().Set("pingFunc", ping())
	// Make the main program not terminate so formatJSON can be called at all.
	select {}
}

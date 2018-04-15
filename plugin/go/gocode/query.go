// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gocode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/nelsam/vidar/suggestion"
)

func query(environ []string, filepath, contents string, runeIndex int) ([]suggestion.Suggestion, error) {
	cmd := exec.Command("gocode", "-f", "json", "autocomplete", filepath, "c"+strconv.Itoa(runeIndex))
	cmd.Env = environ
	cmd.Stdin = bytes.NewBufferString(contents)
	outputJSON, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var output []interface{}
	if err := json.Unmarshal(outputJSON, &output); err != nil {
		return nil, err
	}
	if len(output) < 2 {
		return nil, nil
	}
	completions := output[1].([]interface{})
	if completions[0].(map[string]interface{})["name"].(string) == "PANIC" {
		log.Println("gocode working incorrectly")
		return nil, fmt.Errorf("gocode: invalid output: %+v", output)
	}
	suggestions := make([]suggestion.Suggestion, 0, len(completions))
	for _, completionItem := range completions {
		completion := completionItem.(map[string]interface{})
		suggestions = append(suggestions, suggestion.Suggestion{Name: completion["name"].(string), Signature: completion["type"].(string)})
	}
	return suggestions, nil
}

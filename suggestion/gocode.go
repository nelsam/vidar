// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package suggestion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// FileContainer is any type that contains information about a file.
type FileContainer interface {
	Filepath() string
	Text() string
}

// GoCodeProvider is a gocode-based implementation of gxui.CodeSyntaxProvider.
type GoCodeProvider struct {
	fileContainer FileContainer
	environ       []string
}

func NewGoCodeProvider(fileContainer FileContainer, environ []string) *GoCodeProvider {
	return &GoCodeProvider{
		fileContainer: fileContainer,
		environ:       environ,
	}
}

func (p *GoCodeProvider) SuggestionsAt(runeIndex int) []Suggestion {
	suggestions, err := For(p.environ, p.fileContainer.Filepath(), p.fileContainer.Text(), runeIndex)
	if err != nil {
		log.Printf("Failed to get suggestions: %s", err)
	}
	return suggestions
}

func For(environ []string, filepath, contents string, runeIndex int) ([]Suggestion, error) {
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
	suggestions := make([]Suggestion, 0, len(completions))
	for _, completionItem := range completions {
		completion := completionItem.(map[string]interface{})
		suggestions = append(suggestions, Suggestion{Name: completion["name"].(string), Signature: completion["type"].(string)})
	}
	return suggestions, nil
}

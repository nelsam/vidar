// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package suggestions

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/nelsam/gxui"
)

// FileContainer is any type that contains information about a file.
type FileContainer interface {
	Filepath() string
	Text() string
}

// GoCodeProvider is a gocode-based implementation of gxui.CodeSyntaxProvider.
type GoCodeProvider struct {
	fileContainer FileContainer
	gopath        string
}

func NewGoCodeProvider(fileContainer FileContainer, gopath string) *GoCodeProvider {
	return &GoCodeProvider{
		fileContainer: fileContainer,
		gopath:        gopath,
	}
}

func (p *GoCodeProvider) SuggestionsAt(runeIndex int) []gxui.CodeSuggestion {
	cmd := exec.Command("gocode", "-f", "json", "autocomplete", p.fileContainer.Filepath(), strconv.Itoa(runeIndex))
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH") + ":" + path.Join(p.gopath, "bin"),
		"GOPATH=" + p.gopath,
	}
	cmd.Stdin = bytes.NewBufferString(p.fileContainer.Text())
	outputJSON, err := cmd.Output()
	if err != nil {
		log.Printf("Error: gocode failed: %s", err)
		return nil
	}

	var output []interface{}
	if err := json.Unmarshal(outputJSON, &output); err != nil {
		log.Printf("Error: Could not unmarshal command output as json: %s", err)
		return nil
	}
	if len(output) < 2 {
		return nil
	}
	completions := output[1].([]interface{})
	suggestions := make([]gxui.CodeSuggestion, 0, len(completions))
	for _, completionItem := range completions {
		completion := completionItem.(map[string]interface{})
		suggestions = append(suggestions, suggestion{Value: completion["name"].(string), Type: completion["type"].(string)})
	}
	return suggestions
}

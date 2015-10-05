// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package suggestions

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strconv"

	"github.com/nelsam/gxui"
)

// GoCodeProvider is a gocode-based implementation of gxui.CodeSyntaxProvider.
type GoCodeProvider struct {
	Path   string
	editor gxui.CodeEditor
}

func NewGoCodeProvider(editor gxui.CodeEditor) gxui.CodeSuggestionProvider {
	return &GoCodeProvider{
		editor: editor,
	}
}

func (p *GoCodeProvider) SuggestionsAt(runeIndex int) []gxui.CodeSuggestion {
	cmd := exec.Command("gocode", "-f", "json", "autocomplete", p.Path, strconv.Itoa(runeIndex))
	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Start()
	in.Write([]byte(p.editor.Text()))
	in.Close()
	outputJSON, err := ioutil.ReadAll(out)
	if err != nil {
		panic(err)
	}
	cmd.Wait()

	var output []interface{}
	if err := json.Unmarshal(outputJSON, &output); err != nil {
		panic(err)
	}
	completions := output[1].([]interface{})
	suggestions := make([]gxui.CodeSuggestion, 0, len(completions))
	for _, completionItem := range completions {
		completion := completionItem.(map[string]interface{})
		suggestions = append(suggestions, suggestion{Value: completion["name"].(string), Type: completion["type"].(string)})
	}
	return suggestions
}

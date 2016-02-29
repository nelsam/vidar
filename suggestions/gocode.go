// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package suggestions

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/nelsam/gxui"
)

// GoCodeProvider is a gocode-based implementation of gxui.CodeSyntaxProvider.
type GoCodeProvider struct {
	Path   string
	editor gxui.CodeEditor
	gopath string
}

func NewGoCodeProvider(editor gxui.CodeEditor, gopath string) *GoCodeProvider {
	return &GoCodeProvider{
		editor: editor,
		gopath: gopath,
	}
}

func (p *GoCodeProvider) SuggestionsAt(runeIndex int) []gxui.CodeSuggestion {
	cmd := exec.Command("gocode", "-f", "json", "autocomplete", p.Path, strconv.Itoa(runeIndex))
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH") + ":" + path.Join(p.gopath, "bin"),
		"GOPATH=" + p.gopath,
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Error: Could not access command's STDIN pipe: %s", err)
		return nil
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error: Could not access command's STDOUT pipe: %s", err)
		return nil
	}
	cmd.Start()
	in.Write([]byte(p.editor.Text()))
	in.Close()
	outputJSON, err := ioutil.ReadAll(out)
	if err != nil {
		log.Printf("Error: Could not read command output: %s", err)
		return nil
	}
	cmd.Wait()

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

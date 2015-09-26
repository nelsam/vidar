package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"

	"github.com/google/gxui"
)

type codeSuggestion struct {
	name, code string
}

func (c codeSuggestion) Name() string {
	return c.name
}

func (c codeSuggestion) Code() string {
	return c.code
}

type codeSuggestionProvider struct {
	path   string
	editor gxui.CodeEditor
}

func (s *codeSuggestionProvider) SuggestionsAt(runeIndex int) []gxui.CodeSuggestion {
	cmd := exec.Command("gocode", "-f", "json", "autocomplete", s.path, strconv.Itoa(runeIndex))
	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Start()
	in.Write([]byte(s.editor.Text()))
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
	log.Printf("Output: %v", output)
	completions := output[1].([]interface{})
	log.Printf("Completions: %v", completions)
	suggestions := make([]gxui.CodeSuggestion, 0, len(completions))
	for _, completionItem := range completions {
		completion := completionItem.(map[string]interface{})
		suggestions = append(suggestions, codeSuggestion{name: completion["name"].(string), code: completion["name"].(string)})
	}
	log.Printf("Suggestions: %v", suggestions)
	return suggestions
}

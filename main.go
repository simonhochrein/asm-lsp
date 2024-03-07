package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
	"log"
	"os"
	"strings"
)

//go:embed "6502.json"
var instructions_6502 string

func init() {
	var logger *log.Logger
	defer func() {
		logs.Init(logger)
	}()
	logPath := flag.String("logs", "", "log file path")
	if logPath == nil || *logPath == "" {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	p := *logPath
	f, err := os.Open(p)
	if err == nil {
		logger = log.New(f, "", 0)
		return
	}
	f, err = os.Create(p)
	if err == nil {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	panic(fmt.Sprintf("logs init error: %v", *logPath))
}

func strPtr(s string) *string {
	return &s
}

type InstructionSet struct {
	Instructions []string `json:"instructions"`
}

var files = make(map[defines.DocumentUri]string)

func getCharacterAtPosition(uri defines.DocumentUri, position defines.Position) string {
	file := files[uri]
	lines := strings.Split(file, "\n")
	line := lines[position.Line]
	return line
}

func main() {
	var instructions InstructionSet
	err := json.Unmarshal([]byte(instructions_6502), &instructions)
	if err != nil {
		panic(err)
		return
	}

	server := lsp.NewServer(&lsp.Options{
		CompletionProvider: &defines.CompletionOptions{TriggerCharacters: &[]string{"."}},
	})

	server.OnShutdown(func(ctx context.Context, req *interface{}) error {
		logs.Println("shutdown")
		return nil
	})

	server.OnDidOpenTextDocument(func(ctx context.Context, params *defines.DidOpenTextDocumentParams) error {
		logs.Println("open: ", params)
		files[params.TextDocument.Uri] = params.TextDocument.Text
		return nil
	})
	server.OnHover(func(ctx context.Context, params *defines.HoverParams) (*defines.Hover, error) {
		return &defines.Hover{Contents: "Hello World!"}, nil
	})
	server.OnDidChangeTextDocument(func(ctx context.Context, params *defines.DidChangeTextDocumentParams) error {
		logs.Println("change: ", params)
		files[params.TextDocument.Uri] = params.ContentChanges[0].Text.(string)
		return nil
	})
	server.OnCompletion(func(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
		logs.Println("completion: ", req)
		d := defines.CompletionItemKindClass

		var items []defines.CompletionItem

		line := getCharacterAtPosition(req.TextDocument.Uri, req.Position)[0:req.Position.Character]

		// check if line is preceded with whitespace or a label
		if strings.Contains(line, ";") {
			// comment
		} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || strings.Contains(line, ":") {
			d = defines.CompletionItemKindField

			for _, i := range instructions.Instructions {
				items = append(items, defines.CompletionItem{
					Label: i,
					Kind:  &d,
				})
			}
		} else {
		}

		return &items, nil
	})

	server.OnDocumentFormatting(func(ctx context.Context, req *defines.DocumentFormattingParams) (result *[]defines.TextEdit, err error) {
		logs.Println("format: ", req)

		return &[]defines.TextEdit{}, nil
	})

	server.OnInitialized(func(ctx context.Context, params *defines.InitializeParams) error {
		logs.Println("initialized: ", params)
		return nil
	})

	server.OnInitialize(func(ctx context.Context, params *defines.InitializeParams) (*defines.InitializeResult, *defines.InitializeError) {
		logs.Println("initialize: ", params)

		s := defines.InitializeResult{}
		s.Capabilities.HoverProvider = true
		s.Capabilities.TextDocumentSync = defines.TextDocumentSyncKindFull
		s.Capabilities.CompletionProvider = &defines.CompletionOptions{TriggerCharacters: &[]string{"."}}
		return &s, nil
	})

	server.Run()
}

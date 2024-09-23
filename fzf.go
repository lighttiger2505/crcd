package main

import (
	"strings"
	"sync"

	fzf "github.com/junegunn/fzf/src"
)

type safeString struct {
	s   string
	mux sync.Mutex
}

func (ss *safeString) Store(str string) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	ss.s = str
}

func (ss *safeString) Load() string {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	return ss.s
}

func fzfOpen(inputs []string) (string, error) {
	ss := safeString{}

	inputChan := make(chan string)
	go func() {
		for _, s := range inputs {
			inputChan <- s
		}
		close(inputChan)
	}()

	outputChan := make(chan string)
	go func() {
		for s := range outputChan {
			sp := strings.Split(s, "\n")
			url := sp[1]
			if url != "" {
				ss.Store(url)
			}
		}
	}()

	// Build fzf.Options
	options, err := fzf.ParseOptions(
		true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
		[]string{
			"--ansi",
			"--read0",
			"--multi",
			"--info=inline-right",
			"--reverse",
			"--highlight-line",
			"--cycle",
			"--wrap",
			"--wrap-sign=' ↳ '",
			"--border",
			`--delimiter="\n · "`,
		},
	)
	if err != nil {
		return "", err
	}

	// Set up input and output channels
	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	if _, err := fzf.Run(options); err != nil {
		return "", err
	}
	return ss.Load(), nil
}

package api

import "context"

type matchPrivacyLLM struct {
	prompt string
	calls  int
	after  func()
}

func (m *matchPrivacyLLM) Complete(_ context.Context, prompt string) (string, error) {
	m.calls++
	m.prompt = prompt
	if m.after != nil {
		m.after()
	}
	return "", context.DeadlineExceeded
}

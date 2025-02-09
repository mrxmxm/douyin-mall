package ai

type Client interface {
	Chat(prompt string) (string, error)
}

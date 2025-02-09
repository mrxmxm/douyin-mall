package ai

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) Chat(prompt string) (string, error) {
	return "Mock AI response: " + prompt, nil
}
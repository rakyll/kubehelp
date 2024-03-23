package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

const (
	apiURL = "https://api.anthropic.com/v1/messages"
	model  = "claude-3-opus-20240229"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient(key string) *Client {
	return &Client{
		apiKey: key,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Prompt(prompt string) ([]string, error) {
	data, err := c.readTmplData()
	if err != nil {
		return nil, err
	}
	var systemPrompt bytes.Buffer
	if err := systemPromptTmpl.Execute(&systemPrompt, data); err != nil {
		return nil, err
	}

	resp, err := c.Do(&Request{
		Model:     model,
		MaxTokens: 2048,
		System:    systemPrompt.String(),
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Content) == 0 {
		return nil, nil
	}

	var commands []string
	for _, c := range strings.Split(resp.Content[0].Text, "\n") {
		if c == "" {
			continue
		}
		if !strings.HasPrefix(c, "kubectl") {
			continue
		}
		commands = append(commands, c)
	}
	return commands, nil
}

func (c *Client) Do(r *Request) (*Response, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil
}

func (c *Client) kubectl(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c *Client) readTmplData() (tmplData, error) {
	namespaces, err := c.kubectl("get", "namespaces", "-A")
	if err != nil {
		return tmplData{}, err
	}
	deployments, err := c.kubectl("get", "deployments", "-A")
	if err != nil {
		return tmplData{}, err
	}
	pods, err := c.kubectl("get", "pods", "-A")
	if err != nil {
		return tmplData{}, err
	}
	return tmplData{
		Namespaces:  namespaces,
		Deployments: deployments,
		Pods:        pods,
	}, nil
}

type Request struct {
	Model     string    `json:"model,omitempty"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Response struct {
	Content []Content `json:"content,omitempty"`
}

type Message struct {
	Role    string    `json:"role,omitempty"`
	Content []Content `json:"content,omitempty"`
}

type Content struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type tmplData struct {
	Namespaces  string
	Deployments string
	Pods        string
}

var systemPromptTmpl = template.Must(template.New("prompt").Parse(`
You are a smart Kubernetes command line tool helper.
You will be asked to respond to questions with a kubectl command with appropriate arguments and flags.
Be concise, don't explain your response.
You can suggest more than one kubectl command. List them separately on a new line.
Ensure that you follow user's preferred namespace and other relevant preferences. Once the user starts working in a specific namespace, continue answering for that specific namespace.
Be aware of the current namespaces, services, and pods when answering.
If you cannot find a kubectl command to answer the question, respond with an empty message.
Never explain the response, only respond with lines starting with kubectl.

Existing namespaces available on the cluster:
{{.Namespaces}}

Existing deployments:
{{.Deployments}}

Existing pods:
{{.Pods}}
`))

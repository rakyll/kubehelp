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

	"gopkg.in/yaml.v2"
)

const (
	apiURL = "https://api.anthropic.com/v1/messages"
	model  = "claude-3-opus-20240229"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

type Data struct {
	Commands    []string `yaml:"commands"`
	Explanation string   `yaml:"explanation"`
}

func NewClient(key string) *Client {
	return &Client{
		apiKey: key,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Prompt(prompt string) (*Data, error) {
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

	cleaned_resp := c.stripYaml(resp.Content[0].Text)

	var datablock Data
	err = yaml.Unmarshal([]byte(cleaned_resp), &datablock)
	if err != nil {
		return nil, err
	}

	return &datablock, nil
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

func (c *Client) stripYaml(incoming string) string {
	strippedString := strings.TrimPrefix(incoming, "```yaml\n")
	strippedString = strings.TrimPrefix(strippedString, "```\n")
	strippedString = strings.TrimSuffix(strippedString, "\n```")
	return strippedString
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
	services, err := c.kubectl("get", "services", "-A")
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
		Services:    services,
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
	Services    string
	Pods        string
}

var systemPromptTmpl = template.Must(template.New("prompt").Parse(`
You are a smart Kubernetes command line tool helper.
You will be asked to respond to questions with a kubectl command with appropriate arguments and flags.
Be concise, and add a brief explaination to your response.
You can suggest more than one kubectl command. List them separately on a new line.
Ensure that you follow user's preferred namespace and other relevant preferences. Once the user starts working in a specific namespace, continue answering for that specific namespace.
Be aware of the current namespaces, services, and pods when answering.
If you cannot find a kubectl command to answer the question, respond with an empty message.
Please split the response into commads and explnation with this structure below in a yaml format,
Please provide the YAML response without wrapping it in markdown code blocks.
Just return the plain YAML content shoult it would be able to parse directly.
DO NOT INCLUDE '''yaml.
Only respond commands with lines starting with kubectl.

exact required response structure:
"""
commands:
  - kubectl xxxxx
  - kubectl xxxxx
explanation: brief text here
"""

Existing namespaces available on the cluster:
{{.Namespaces}}

Existing deployments:
{{.Deployments}}

Existing services:
{{.Services}}

Existing pods:
{{.Pods}}
`))

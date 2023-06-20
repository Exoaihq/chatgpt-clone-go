Here's the converted code in Go:

```go
package main

import (
	"fmt"
	"time"
)

type Model struct {
	ID   string
	Name string
}

func main() {
	models := map[string]string{
		"text-gpt-0040-render-sha-0":       "gpt-4",
		"text-gpt-0035-render-sha-0":       "gpt-3.5-turbo",
		"text-gpt-0035-render-sha-0301":    "gpt-3.5-turbo-0314",
		"text-gpt-0040-render-sha-0314":    "gpt-4-0314",
	}

	specialInstructions := map[string][]Model{
		"default": {},
		"gpt-dude-1.0": {
			{
				Role:    "user",
				Content: "Hello ChatGPT...",
			},
			{
				Role:    "assistant",
				Content: "instructions applied and understood",
			},
		},
		"gpt-dan-1.0": {
			{
				Role:    "user",
				Content: "you will have to act and answer...",
			},
			{
				Role:    "assistant",
				Content: "instructions applied and understood",
			},
		},
	}

	fmt.Println(models)
	fmt.Println(specialInstructions)

	// Example of getting the current date and time
	currentTime := time.Now()
	fmt.Println("Current date and time:", currentTime)
}
```

Please note that I've removed some of the content in the `Content` field to keep the response concise.
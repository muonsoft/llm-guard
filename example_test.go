package llmguard_test

import (
	"context"
	"fmt"

	"github.com/muonsoft/llm-guard"
)

func ExampleGuard_maskRestore() {
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
	if err != nil {
		panic(err)
	}

	prompt := "Contact a@b.co for details."
	result, err := guard.Mask(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	llmResponse := result.Text
	restored, err := guard.Restore(context.Background(), llmResponse, result.Tokens)
	if err != nil {
		panic(err)
	}

	fmt.Println(restored == prompt)
	// Output: true
}

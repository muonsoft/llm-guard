package llmguard_test

import (
	"context"
	"fmt"

	"github.com/muonsoft/llm-guard"
)

func ExampleGuard_structuredPack() {
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewPhoneDetector()),
		llmguard.WithDetector(llmguard.NewIPDetector()),
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewINNDetector()),
		llmguard.WithDetector(llmguard.NewSNILSDetector()),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
	)
	if err != nil {
		panic(err)
	}

	prompt := "Contact +7 999 123 45 67 or https://example.com"
	result, err := guard.Mask(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	if err != nil {
		panic(err)
	}

	fmt.Println(restored == prompt)
	// Output: true
}

func ExampleGuard_customRegexpEntity() {
	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-[0-9]{6}`,
		Confidence: 0.9,
	})
	if err != nil {
		panic(err)
	}

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	if err != nil {
		panic(err)
	}

	prompt := "Assign EMP-123456 to the project."
	result, err := guard.Mask(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	if err != nil {
		panic(err)
	}

	fmt.Println(restored == prompt)
	// Output: true
}

func ExampleGuard_personMaskRestore() {
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewPersonDetector()))
	if err != nil {
		panic(err)
	}

	prompt := "Документ подписал Иван Петров."
	result, err := guard.Mask(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	if err != nil {
		panic(err)
	}

	fmt.Println(restored == prompt)
	// Output: true
}

func ExampleGuard_addressMaskRestore() {
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewAddressDetector()))
	if err != nil {
		panic(err)
	}

	prompt := "Доставка: ул. Ленина, 10."
	result, err := guard.Mask(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	if err != nil {
		panic(err)
	}

	fmt.Println(restored == prompt)
	// Output: true
}

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

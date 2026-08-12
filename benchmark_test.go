package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
)

var (
	benchGuard       *llmguard.Guard
	benchRUInput     string
	benchMixedInput  string
	benchSecretInput string
	benchMaskedText  string
	benchTokenSet    *llmguard.TokenSet
)

func init() {
	var err error
	benchGuard, err = llmguard.New(
		llmguard.WithDetector(llmguard.NewPersonDetector()),
		llmguard.WithDetector(llmguard.NewAddressDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPhoneDetector()),
		llmguard.WithDetector(llmguard.NewIPDetector()),
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewINNDetector()),
		llmguard.WithDetector(llmguard.NewSNILSDetector()),
		llmguard.WithDetector(llmguard.NewPassportDetector()),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
		llmguard.WithDetector(llmguard.NewBankAccountDetector()),
		llmguard.WithDetector(llmguard.NewDateOfBirthDetector()),
		llmguard.WithDetector(llmguard.NewJWTDetector()),
		llmguard.WithDetector(llmguard.NewPEMPrivateKeyDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithDetector(llmguard.NewDSNDetector()),
		llmguard.WithSecretAction(llmguard.ActionMask),
	)
	if err != nil {
		panic(err)
	}

	benchRUInput = "Документ подписал Иван Петров. Доставка: ул. Ленина, 10."
	benchMixedInput = "Контакт +7 999 123 45 67, a@b.co, ИНН 7707083893"
	benchSecretInput = "token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJzeW50aGV0aWMifQ.c2ln"

	result, err := benchGuard.Mask(context.Background(), benchMixedInput)
	if err != nil {
		panic(err)
	}
	benchMaskedText = result.Text
	benchTokenSet = result.Tokens
}

func BenchmarkDetect_RUPrompt(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Detect(ctx, benchRUInput)
	}
}

func BenchmarkDetect_MixedPII(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Detect(ctx, benchMixedInput)
	}
}

func BenchmarkDetect_SyntheticSecret(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Detect(ctx, benchSecretInput)
	}
}

func BenchmarkMask_RUPrompt(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Mask(ctx, benchRUInput)
	}
}

func BenchmarkMask_MixedPII(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Mask(ctx, benchMixedInput)
	}
}

func BenchmarkMask_SyntheticSecret(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Mask(ctx, benchSecretInput)
	}
}

func BenchmarkRestore_MixedPII(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Restore(ctx, benchMaskedText, benchTokenSet)
	}
}

func BenchmarkRestore_RUPrompt(b *testing.B) {
	ctx := context.Background()
	result, err := benchGuard.Mask(ctx, benchRUInput)
	if err != nil {
		b.Fatal(err)
	}
	text := result.Text
	tokens := result.Tokens
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Restore(ctx, text, tokens)
	}
}

func BenchmarkRestore_SyntheticSecret(b *testing.B) {
	ctx := context.Background()
	result, err := benchGuard.Mask(ctx, benchSecretInput)
	if err != nil {
		b.Fatal(err)
	}
	text := result.Text
	tokens := result.Tokens
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchGuard.Restore(ctx, text, tokens)
	}
}

func BenchmarkObserver_DefaultNoop(b *testing.B) {
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = guard.Detect(ctx, "mail a@b.co")
	}
}

func BenchmarkObserver_WithCallback(b *testing.B) {
	var sink [1]byte
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(llmguard.ObserverFunc(func(event llmguard.Event) {
			sink[0] = event.Operation[0]
		})),
	)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = guard.Detect(ctx, "mail a@b.co")
	}
}

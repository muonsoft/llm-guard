package nlp

import "context"

type cancelAfterKey struct{}

type cancelAfterFunc func(step string) error

func withCancelAfter(ctx context.Context, fn cancelAfterFunc) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, cancelAfterKey{}, fn)
}

func checkScanCtx(ctx context.Context, step string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if fn, ok := ctx.Value(cancelAfterKey{}).(cancelAfterFunc); ok && fn != nil {
		if err := fn(step); err != nil {
			return err
		}
	}
	return nil
}

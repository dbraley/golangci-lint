//args: -Ectxlint
package testdata

import "context"

func takesCtx(ctx context.Context, number int)                      {}
func ignoresCtx(_ context.Context, number int)                      {}
func takesCtxNotFirst(name string, ctx context.Context, number int) {} // ERROR "Context should be the first parameter"
func takesCtxTwice(ctx, _ context.Context, number int)              {} // ERROR "Context should be the first parameter"
func takesCtxWrongName(cntxt context.Context, number int)           {} // ERROR "Context parameter should be called 'ctx'"

func caller() {
	ctx := context.Background()         // ERROR "do not use context.Background"
	takesCtx(context.Background(), 123) // ERROR "do not use context.Background"
	takesCtx(ctx, 123)
}

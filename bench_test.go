package alloter

import (
	"context"
	"testing"
)

func Benchmark_Block(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		select {
		case <-ctx.Done():
			ctx.Err()
		default:

		}
	}

}

func Benchmark_For(b *testing.B) {
	for i := 0; i < b.N; i++ {
	}
}

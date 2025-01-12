package ringbuffer

import "testing"

func BenchmarkRingBuffer(b *testing.B) {
	r := NewRingBuffer[int](uint64(b.N))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.PushBack(i)
		_, _ = r.PopFront()
	}
}

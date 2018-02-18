package mlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type discardSliceWriter struct{}

func (d *discardSliceWriter) WriteString(s string) (int, error) { return len(s), nil }
func (d *discardSliceWriter) Write(b []byte) (int, error)       { return len(b), nil }
func (d *discardSliceWriter) WriteByte(c byte) error            { return nil }
func (d *discardSliceWriter) Truncate(i int)                    {}

func BenchmarkLogMapUnsortedWriteBuf(b *testing.B) {
	buf := &discardSliceWriter{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(10, false)] = randString(25, true)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.unsortedWriteBuf(buf)
		buf.Truncate(0)
	}
}

func BenchmarkLogMapSortedWriteBuf(b *testing.B) {
	buf := &discardSliceWriter{}
	m := Map{}
	for i := 1; i <= 100; i++ {
		m[randString(10, false)] = randString(25, true)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.sortedWriteBuf(buf)
		buf.Truncate(0)
	}
}

func TestLogMapWriteTo(t *testing.T) {
	m := Map{"test": "this is \"a test\" of \t some \n a"}
	buf := &sliceBuffer{make([]byte, 0, 1024)}
	m.sortedWriteBuf(buf)
	n := `test="this is \"a test\" of \t some \n a"`
	l := buf.String()
	assert.Equal(t, n, l, "did not match")

}

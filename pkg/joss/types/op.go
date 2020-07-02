package types

import "context"

type Op struct {
	Bucket      string
	ThreadCount int
	PartSize    int64

	MaxKeys   int64
	Delimiter string
	Context   context.Context

	Metadata map[string]*string

	ProgressFn ProgressFn
}

type ProgressFn func(totalBytes int64, consumedBytes int64)
type OpOption func(*Op)

func (op *Op) ApplyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

func WithBucket(bucket string) OpOption {
	return func(op *Op) { op.Bucket = bucket }
}

func WithProgress(fn ProgressFn) OpOption {
	return func(op *Op) {
		op.ProgressFn = fn
	}
}

func WithMaxKeys(maxkeys int64) OpOption {
	return func(op *Op) { op.MaxKeys = maxkeys }
}

// 分片大小单位字节,最低5mb
func WithPartSize(partSize int64) OpOption {
	return func(op *Op) {
		op.PartSize = partSize
	}
}

func WithThreadCount(count int) OpOption {
	return func(op *Op) { op.ThreadCount = count }
}

func WithDelimiter() OpOption {
	return func(op *Op) {
		op.Delimiter = "/"
	}
}

func WithContext(ctx context.Context) OpOption {
	return func(op *Op) { op.Context = ctx }
}

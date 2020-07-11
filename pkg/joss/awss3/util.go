package awss3

import (
	"github.com/mayongze/joss-cli/pkg/joss/types"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type readerAtSeeker interface {
	io.ReaderAt
	io.ReadSeeker
}

type CustomReader struct {
	readerAtSeeker readerAtSeeker
	consumedBytes  int64
	totalBytes     int64

	progressFn   types.ProgressFn
	offMap       map[int64]struct{}
	mux          sync.Mutex
	progressTick *time.Ticker
	once         sync.Once
}

func NewCustomReader(readerAtSeeker readerAtSeeker, progressFn types.ProgressFn) *CustomReader {
	n, _ := readerAtSeeker.Seek(0, io.SeekEnd)
	_, _ = readerAtSeeker.Seek(0, io.SeekStart)
	if progressFn == nil {
		progressFn = func(totalBytes int64, consumedBytes int64) {}
	}
	return &CustomReader{
		readerAtSeeker: readerAtSeeker,
		consumedBytes:  0,
		totalBytes:     n,
		progressFn:     progressFn,
		offMap:         map[int64]struct{}{},
	}
}

func (r *CustomReader) Read(p []byte) (int, error) {
	return r.readerAtSeeker.Read(p)
}

func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.readerAtSeeker.Seek(offset, whence)
}

func (r *CustomReader) publishLoop() {
	// r.progressFn(r.totalBytes, r.consumedBytes)
	for range r.progressTick.C {
		r.progressFn(r.totalBytes, r.consumedBytes)
	}
}

func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	// 会调用2次因为会预先计算一次md5, 然后在上传 10485760 32768 15728640 20971520 10518528 65536 10551296 98304 10584064
	n, err := r.readerAtSeeker.ReadAt(p, off)
	if err != nil {
		return n, err
	}
	// 同一个位置点第二次出现的时候开始计算size
	r.mux.Lock()
	r.once.Do(func() {
		r.progressTick = time.NewTicker(time.Millisecond * 800)
		go r.publishLoop()
	})
	if _, ok := r.offMap[off]; ok {
		atomic.AddInt64(&r.consumedBytes, int64(n))
		if r.totalBytes <= r.consumedBytes {
			r.progressFn(r.totalBytes, r.consumedBytes)
			r.Close()
		}
	} else {
		r.offMap[off] = struct{}{}
	}
	r.mux.Unlock()
	// Got the length have read( or means has uploaded), and you can construct your message

	// I have no idea why the read length need to be div 2,
	// maybe the request read once when Sign and actually send call ReadAt again
	// It works for me
	//log.Printf("Current upload file：%s",r.f_info.Key)
	return n, err
}

func (r *CustomReader) Close() {
	r.progressTick.Stop()
}

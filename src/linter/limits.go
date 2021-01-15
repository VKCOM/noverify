package linter

import "github.com/VKCOM/noverify/src/lintdebug"

// parseWaiter waits to allow parsing of a file.
type parseWaiter struct {
	size int
}

type memoryRequest struct {
	size     int
	filename string
}

var (
	parseStartCh    = make(chan memoryRequest)
	parseFinishedCh = make(chan int)
)

// MemoryLimiterThread starts memory limiter goroutine that disallows to use parse files more than MaxFileSize
// total bytes.
func MemoryLimiterThread() {
	var used int

	plusCh := parseStartCh
	minusCh := parseFinishedCh

	for {
		select {
		case req := <-plusCh:
			used += req.size
			if used > MaxFileSize {
				lintdebug.Send("Limiting concurrency to save memory: currently parsing %s, total file size %d KiB", req.filename, used/1024)
				plusCh = nil
			}
		case sz := <-minusCh:
			used -= sz
			if used <= MaxFileSize {
				plusCh = parseStartCh
			}
		}
	}
}

// beforeParse must be called before parsing file, so that soft memory
// limit can be applied.
// Do not forget to call Finish()!
func beforeParse(size int, filename string) *parseWaiter {
	parseStartCh <- memoryRequest{size: size, filename: filename}
	return &parseWaiter{
		size: size,
	}
}

// Finish must be called after parsing is finished (e.g. using defer p.Finish()) to
// allow other goroutines to parse files.
func (p *parseWaiter) Finish() {
	parseFinishedCh <- p.size
}

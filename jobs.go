package geard

import (
	"io"
)

type Job interface {
	Execute()

	Fast() bool
	JobId() RequestIdentifier
}

type Join interface {
	Join(Job, <-chan bool) (bool, <-chan bool, error)
}

// A job may return a structured error, a stream of unstructured data,
// or a stream of structured data.  In general, jobs only stream on
// success - a failure is written immediately.  A streaming job
// may write speculative side channel data that will be returned when
// a successful response occurs, or thrown away when an error is written.
// Error writes are final
type JobResponse interface {
	StreamResult() bool

	Success(t JobResponseSuccess)
	SuccessWithData(t JobResponseSuccess, data interface{})
	SuccessWithWrite(t JobResponseSuccess, flush bool) io.Writer
	Failure(reason JobError)

	WriteClosed() <-chan bool
	WritePendingSuccess(name string, value interface{})
}

type JobResponseSuccess int
type JobResponseFailure int

// A structured error response for a job.
type JobError interface {
	error
	ResponseFailure() JobResponseFailure
	ResponseData() interface{} // May be nil if no data is returned to a client
}

type jobRequest struct {
	RequestId RequestIdentifier
}

func (j *jobRequest) Fast() bool {
	return false
}

func (j *jobRequest) JobId() RequestIdentifier {
	return j.RequestId
}

/*
Package errgo is a small lib for stacked errors


Copyright (c) 2013, Dan Kortschak
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
Neither the name of the <ORGANIZATION> nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package errgo

// Diagnosis is an error and layered error annotations.
type Diagnosis interface {
	// The error behavior of a Diagnosis is based on the last annotation applied.
	error

	Cause() error               // Cause returns the initial error in the Diagnosis.
	Wrap(error) Diagnosis       // Wrap adds an annotation layer to the Diagnosis.
	Unwrap() (Diagnosis, error) // Unwrap returns the Diagnosis and the most recent annotation.
}

// AllUnwrapper is an optional interface used by the UnwrapAll function.
type AllUnwrapper interface {
	UnwrapAll() []error // UnwrapAll returns a flat list of errors in order of annotation.
}

// New returns a new Diagnosis based on the provided error. If the error is a Diagnosis it
// is returned unaltered.
func New(err error) Diagnosis {
	if d, ok := err.(Diagnosis); ok {
		return d
	}
	return diagnosis{err}
}

// Cause returns the initially identified cause of an error if the error is a Diagnosis, or the error
// itself if it is not.
func Cause(err error) error {
	if d, ok := err.(Diagnosis); ok {
		return d.Cause()
	}
	return err
}

// Wrap adds an annotation to an error, returning a Diagnosis.
func Wrap(err, annotation error) Diagnosis { return New(err).Wrap(annotation) }

// Unwrap returns the most recent annotation of an error and the remaining diagnosis
// after the annotation is removed or nil if no further errors remain. Unwrap returns
// a nil Diagnosis if the error is not a Diagnosis.
func Unwrap(err error) (Diagnosis, error) {
	if d, ok := err.(Diagnosis); ok {
		return d.Unwrap()
	}
	return nil, err
}

// UnwrapAll returns a flat list of errors in order of annotation. If the provided error is
// not a Diagnosis a single element slice of error is returned containing the error.
func UnwrapAll(err error) []error {
	if err == nil {
		return nil
	}
	switch d := err.(type) {
	case AllUnwrapper:
		return d.UnwrapAll()
	case Diagnosis:
		var errs []error
		for d != nil {
			d, err = d.Unwrap()
			errs = append(errs, err)
		}
		return reverse(errs)
	default:
		return []error{err}
	}
}

func reverse(err []error) []error {
	for i, j := 0, len(err)-1; i < j; i, j = i+1, j-1 {
		err[i], err[j] = err[j], err[i]
	}
	return err
}

// diagnosis is the basic implementation.
type diagnosis []error

func (d diagnosis) Error() string {
	if len(d) > 0 {
		return d[len(d)-1].Error()
	}
	return ""
}
func (d diagnosis) Cause() error {
	if len(d) > 0 {
		return d[0]
	}
	return nil
}
func (d diagnosis) Wrap(err error) Diagnosis { return append(d, err) }
func (d diagnosis) Unwrap() (Diagnosis, error) {
	switch len(d) {
	case 0:
		return nil, nil
	case 1:
		return nil, d[0]
	default:
		return d[:len(d)-1], d[len(d)-1]
	}
}

func (d diagnosis) UnwrapAll() []error { return d }

package io

type constError string

func (err constError) Error() string {
	return string(err)
}

type Record []string

const (
	ErrOpeningFile        = constError("error opening file")
	ErrInvalidRecord      = constError("invalid record found")
	ErrInvalidWriteRecord = constError("invalid writing of record")
)

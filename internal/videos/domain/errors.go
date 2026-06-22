package domain

import "fmt"

type VideoError interface {
	error
	isVideoError()
}

type VideoNotFoundError struct {
	ID string
}

func (e *VideoNotFoundError) Error() string { return "Video not found: " + e.ID }
func (e *VideoNotFoundError) isVideoError() {}

type VideoInvalidStatusTransitionError struct {
	Current VideoStatus
	Target  VideoStatus
}

func (e *VideoInvalidStatusTransitionError) Error() string {
	return fmt.Sprintf("Invalid status transition from %s to %s", e.Current, e.Target)
}
func (e *VideoInvalidStatusTransitionError) isVideoError() {}

type VideoRawNotFoundError struct {
	ID string
}

func (e *VideoRawNotFoundError) Error() string {
	return "Raw video file not found in storage for video: " + e.ID
}
func (e *VideoRawNotFoundError) isVideoError() {}

type VideoNotReadyError struct {
	ID string
}

func (e *VideoNotReadyError) Error() string {
	return "Video is not ready for streaming: " + e.ID
}
func (e *VideoNotReadyError) isVideoError() {}

type VideoInvalidTypeError struct {
	Type string
}

func (e *VideoInvalidTypeError) Error() string {
	return "Invalid video type: " + e.Type
}
func (e *VideoInvalidTypeError) isVideoError() {}

package domain

import (
	"time"

	shareddomain "go-starter/internal/shared/domain"
)

type VideoQuality string

const (
	Quality480p  VideoQuality = "480p"
	Quality720p  VideoQuality = "720p"
	Quality1080p VideoQuality = "1080p"
	Quality4k    VideoQuality = "4k"
)

type VideoStatus string

const (
	StatusPendingUpload VideoStatus = "PendingUpload"
	StatusProcessing    VideoStatus = "Processing"
	StatusReady         VideoStatus = "Ready"
	StatusFailed        VideoStatus = "Failed"
	StatusReplacing     VideoStatus = "Replacing"
)

var validTransitions = map[VideoStatus][]VideoStatus{
	StatusPendingUpload: {StatusProcessing},
	StatusProcessing:    {StatusReady, StatusFailed},
	StatusReady:         {StatusReplacing},
	StatusReplacing:     {StatusPendingUpload},
	StatusFailed:        {StatusProcessing},
}

func (s VideoStatus) CanTransitionTo(target VideoStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

func (s VideoStatus) String() string { return string(s) }

type VideoType string

const (
	VideoTypeMovie       VideoType = "movie"
	VideoTypeSeries      VideoType = "series"
	VideoTypeDocumentary VideoType = "documentary"
)

func (t VideoType) IsValid() bool {
	switch t {
	case VideoTypeMovie, VideoTypeSeries, VideoTypeDocumentary:
		return true
	default:
		return false
	}
}

func (t VideoType) String() string { return string(t) }

type Video struct {
	ID          shareddomain.Id
	Title       string
	Description string
	Type        VideoType
	Status      VideoStatus
	Qualities   []VideoQuality
	RawPath     string
	PhotoPath   string
	UploadedAt  time.Time
}

func NewVideo(id shareddomain.Id, title, description string, videoType VideoType) *Video {
	idStr := id.String()
	return &Video{
		ID:          id,
		Title:       title,
		Description: description,
		Type:        videoType,
		Status:      StatusPendingUpload,
		Qualities:   []VideoQuality{},
		RawPath:     "raws/" + idStr + ".mp4",
		PhotoPath:   "photos/" + idStr + "/thumbnail.jpg",
		UploadedAt:  time.Now().UTC(),
	}
}

func (v *Video) GetID() string                { return v.ID.String() }
func (v *Video) GetTitle() string             { return v.Title }
func (v *Video) GetDescription() string       { return v.Description }
func (v *Video) GetType() VideoType           { return v.Type }
func (v *Video) GetStatus() VideoStatus       { return v.Status }
func (v *Video) GetQualities() []VideoQuality { return v.Qualities }
func (v *Video) GetRawPath() string           { return v.RawPath }
func (v *Video) GetPhotoPath() string         { return v.PhotoPath }
func (v *Video) GetUploadedAt() time.Time     { return v.UploadedAt }

func (v *Video) TransitionTo(target VideoStatus) error {
	if !v.Status.CanTransitionTo(target) {
		return &VideoInvalidStatusTransitionError{
			Current: v.Status,
			Target:  target,
		}
	}
	v.Status = target
	return nil
}

func (v *Video) SetQualities(qualities []VideoQuality) {
	v.Qualities = qualities
}

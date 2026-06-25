package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"go-starter/internal/videos/domain"
)

type FfmpegAdapter struct{}

func NewFfmpegAdapter() *FfmpegAdapter {
	return &FfmpegAdapter{}
}

func (a *FfmpegAdapter) TranscodeToHLS(ctx context.Context, params domain.TranscodeParams) error {
	scale, maxrate, bufsize := qualityToParams(params.Quality)
	if scale == "" {
		return fmt.Errorf("unsupported quality: %s", params.Quality)
	}

	segmentPattern := filepath.Join(params.OutputDir, "seg-%d.ts")
	playlistPath := filepath.Join(params.OutputDir, "playlist.m3u8")

	args := []string{
		"-i", params.InputPath,
		"-vf", scale,
		"-c:v", "libx264",
		"-preset", "fast",
		"-g", "180",
		"-keyint_min", "180",
		"-maxrate", maxrate,
		"-bufsize", bufsize,
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg %s: %w\n%s", params.Quality, err, stderr.String())
	}

	return nil
}

func qualityToParams(quality string) (scale, maxrate, bufsize string) {
	switch quality {
	case "480p":
		return "scale=-2:480", "800k", "1600k"
	case "720p":
		return "scale=-2:720", "2800k", "5600k"
	case "1080p":
		return "scale=-2:1080", "5000k", "10000k"
	default:
		return "", "", ""
	}
}

package application

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-starter/internal/videos/domain"
)

type ProcessVideoInput struct {
	VideoID            string
	RequestedQualities []string
	IsAppend           bool
}

type ProcessVideo struct {
	repo       domain.IVideoRepository
	storage    domain.IStorageAdapter
	transcoder domain.ITranscoder
}

func NewProcessVideo(
	repo domain.IVideoRepository,
	storage domain.IStorageAdapter,
	transcoder domain.ITranscoder,
) *ProcessVideo {
	return &ProcessVideo{
		repo:       repo,
		storage:    storage,
		transcoder: transcoder,
	}
}

func (uc *ProcessVideo) Execute(ctx context.Context, input ProcessVideoInput) error {
	slog.Info("processing video", "video_id", input.VideoID, "qualities", input.RequestedQualities, "is_append", input.IsAppend)

	workDir, err := os.MkdirTemp("", "video-worker-*")
	if err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	rawPath := filepath.Join(workDir, "input.mp4")
	rawKey := "raws/" + input.VideoID + ".mp4"
	if err := uc.downloadRaw(ctx, rawKey, rawPath); err != nil {
		return uc.failVideo(ctx, input.VideoID, fmt.Errorf("download raw: %w", err))
	}

	var existingMaster []byte
	if input.IsAppend {
		existingMaster, err = uc.downloadMaster(ctx, input.VideoID)
		if err != nil {
			return uc.failVideo(ctx, input.VideoID, fmt.Errorf("download master playlist: %w", err))
		}
	}

	for _, quality := range input.RequestedQualities {
		qualityDir := filepath.Join(workDir, quality)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			return uc.failVideo(ctx, input.VideoID, fmt.Errorf("create quality dir %s: %w", quality, err))
		}

		if err := uc.transcoder.TranscodeToHLS(ctx, domain.TranscodeParams{
			InputPath: rawPath,
			Quality:   quality,
			OutputDir: qualityDir,
		}); err != nil {
			return uc.failVideo(ctx, input.VideoID, fmt.Errorf("transcode %s: %w", quality, err))
		}

		if err := uc.uploadDir(ctx, input.VideoID, quality, qualityDir); err != nil {
			return uc.failVideo(ctx, input.VideoID, fmt.Errorf("upload %s: %w", quality, err))
		}
	}

	masterContent := buildMasterPlaylist(input, existingMaster)
	masterKey := "videos/" + input.VideoID + "/master.m3u8"
	if _, err := uc.storage.UploadFile(ctx, masterKey, masterContent, "application/vnd.apple.mpegurl"); err != nil {
		return uc.failVideo(ctx, input.VideoID, fmt.Errorf("upload master: %w", err))
	}

	video, err := uc.repo.FindByID(ctx, input.VideoID)
	if err != nil || video == nil {
		return uc.failVideo(ctx, input.VideoID, fmt.Errorf("find video: %w", err))
	}

	if input.IsAppend {
		video.SetQualities(mergeQualities(video.GetQualities(), input.RequestedQualities))
	} else {
		video.SetQualities(stringsToDomainQualities(input.RequestedQualities))
	}

	if err := video.TransitionTo(domain.StatusReady); err != nil {
		return uc.failVideo(ctx, input.VideoID, err)
	}

	if _, err := uc.repo.Update(ctx, video); err != nil {
		return uc.failVideo(ctx, input.VideoID, fmt.Errorf("update video: %w", err))
	}

	slog.Info("video processing complete", "video_id", input.VideoID)
	return nil
}

func (uc *ProcessVideo) downloadRaw(ctx context.Context, key, destPath string) error {
	url, err := uc.storage.GeneratePresignedGetUrl(ctx, key, 10*time.Minute)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("copy body: %w", err)
	}

	return nil
}

func (uc *ProcessVideo) downloadMaster(ctx context.Context, videoID string) ([]byte, error) {
	key := "videos/" + videoID + "/master.m3u8"
	url, err := uc.storage.GeneratePresignedGetUrl(ctx, key, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get master: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("master download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (uc *ProcessVideo) uploadDir(ctx context.Context, videoID, quality, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	prefix := "videos/" + videoID + "/" + quality + "/"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		key := prefix + entry.Name()
		contentType := "video/mp2t"
		if strings.HasSuffix(entry.Name(), ".m3u8") {
			contentType = "application/vnd.apple.mpegurl"
		}

		if _, err := uc.storage.UploadFile(ctx, key, data, contentType); err != nil {
			return fmt.Errorf("upload %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (uc *ProcessVideo) failVideo(ctx context.Context, videoID string, err error) error {
	slog.Error("video processing failed", "video_id", videoID, "error", err)

	video, findErr := uc.repo.FindByID(ctx, videoID)
	if findErr != nil || video == nil {
		return err
	}

	video.TransitionTo(domain.StatusFailed)
	uc.repo.Update(ctx, video)
	return err
}

func buildMasterPlaylist(input ProcessVideoInput, existingMaster []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("#EXTM3U\n")

	if input.IsAppend && len(existingMaster) > 0 {
		lines := strings.Split(string(existingMaster), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "#EXTM3U" {
				buf.WriteString(line + "\n")
			}
		}
	}

	for _, quality := range input.RequestedQualities {
		switch quality {
		case "480p":
			buf.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=854x480\n")
		case "720p":
			buf.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720\n")
		case "1080p":
			buf.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080\n")
		}
		buf.WriteString(quality + "/playlist.m3u8\n")
	}

	return buf.Bytes()
}

func mergeQualities(existing []domain.VideoQuality, newQ []string) []domain.VideoQuality {
	seen := make(map[string]bool)
	for _, q := range existing {
		seen[string(q)] = true
	}
	for _, q := range newQ {
		if !seen[q] {
			existing = append(existing, domain.VideoQuality(q))
			seen[q] = true
		}
	}
	return existing
}

func stringsToDomainQualities(qualities []string) []domain.VideoQuality {
	res := make([]domain.VideoQuality, len(qualities))
	for i, q := range qualities {
		res[i] = domain.VideoQuality(q)
	}
	return res
}

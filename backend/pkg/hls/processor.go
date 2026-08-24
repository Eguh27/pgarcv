package hls

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Processor struct {
	OutputBase string
	EncryptKey string
	ServerURL  string
}

func NewProcessor(outputBase, encryptKey, serverURL string) *Processor {
	return &Processor{OutputBase: outputBase, EncryptKey: encryptKey, ServerURL: serverURL}
}

func (p *Processor) ExtractThumbnail(videoPath string, atSecond int) (string, error) {
	thumbDir := filepath.Join(p.OutputBase, "thumbnails")
	os.MkdirAll(thumbDir, 0755)
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	thumbPath := filepath.Join(thumbDir, base+"_thumb.jpg")
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", strconv.Itoa(atSecond),
		"-vframes", "1",
		"-vf", "scale=1280:720:force_original_aspect_ratio=decrease",
		"-q:v", "2",
		"-an",
		"-y",
		thumbPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg thumbnail: %s — %w", string(out), err)
	}
	return thumbPath, nil
}

func (p *Processor) GetDuration(videoPath string) (int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return int(dur), nil
}

func (p *Processor) ProcessHLS(videoID uint, videoPath string) (string, error) {
	hlsDir := filepath.Join(p.OutputBase, "hls", fmt.Sprintf("video_%d", videoID))
	os.MkdirAll(hlsDir, 0755)

	m3u8Path := filepath.Join(hlsDir, "index.m3u8")

	args := []string{
		"-i", videoPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-ac", "2",
		"-movflags", "+faststart",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(hlsDir, "seg_%03d.ts"),
		"-hls_flags", "independent_segments",
		"-y", m3u8Path,
	}

	// Enkripsi hanya kalau HLS_ENCRYPT_KEY ada
	if p.EncryptKey != "" {
		keyBytes, err := hex.DecodeString(p.EncryptKey)
		if err != nil || len(keyBytes) != 16 {
			return "", fmt.Errorf("HLS_ENCRYPT_KEY harus 32 hex chars")
		}
		keyPath := filepath.Join(hlsDir, "enc.key")
		keyInfoPath := filepath.Join(hlsDir, "enc.keyinfo")
		if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
			return "", err
		}
		keyURL := fmt.Sprintf("%s/api/hls/key/%d", p.ServerURL, videoID)
		if err := os.WriteFile(keyInfoPath, []byte(fmt.Sprintf("%s\n%s\n", keyURL, keyPath)), 0600); err != nil {
			return "", err
		}
		// sisipkan flag key info sebelum output m3u8 (elemen terakhir adalah "-y" dan m3u8Path)
		// cari index "-y" terakhir
		lastY := -1
		for i := len(args) - 1; i >= 0; i-- {
			if args[i] == "-y" {
				lastY = i
				break
			}
		}
		if lastY == -1 {
			return "", fmt.Errorf("internal error: ffmpeg args tidak punya flag -y")
		}
		before := append([]string{}, args[:lastY]...)
		after := append([]string{}, args[lastY:]...)
		args = append(before, "-hls_key_info_file", keyInfoPath)
		args = append(args, after...)

	}

	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg HLS: %s — %w", string(out), err)
	}
	return hlsDir, nil
}

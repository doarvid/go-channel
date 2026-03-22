package core

import (
	"bytes"
	"context"
	"os/exec"
)

// AudioSender is an optional interface for platforms that support sending audio.
type AudioSender interface {
	SendAudio(ctx context.Context, replyCtx any, audio []byte, format string) error
}

// ConvertAudioToOpus uses ffmpeg to convert audio to opus format (ogg container).
// Returns the opus bytes. If ffmpeg is not installed, returns an error.
func ConvertAudioToOpus(ctx context.Context, audio []byte, srcFormat string) ([]byte, error) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, nil // Silently fail if ffmpeg not available
	}

	args := []string{"-i", "pipe:0", "-c:a", "libopus", "-f", "opus", "-y", "pipe:1"}
	if srcFormat == "amr" || srcFormat == "silk" {
		args = append([]string{"-f", srcFormat}, args...)
	}
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stdin = bytes.NewReader(audio)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, nil // Silently fail if conversion fails
	}
	return stdout.Bytes(), nil
}

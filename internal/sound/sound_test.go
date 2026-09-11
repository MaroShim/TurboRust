package sound

import (
	"bytes"
	"os"
	"testing"
)

func TestGenerateWarmConeTone(t *testing.T) {
	tones := []tone{
		{freq: 440, durationMs: 50},
		{freq: 880, durationMs: 50},
	}
	wavData := generateWarmConeTone(tones)

	if len(wavData) <= 44 {
		t.Fatalf("expected WAV data to be longer than header (44 bytes), got %d bytes", len(wavData))
	}

	// Verify RIFF / WAVE header markers
	if !bytes.Equal(wavData[0:4], []byte("RIFF")) {
		t.Errorf("expected RIFF header marker, got %s", string(wavData[0:4]))
	}
	if !bytes.Equal(wavData[8:12], []byte("WAVE")) {
		t.Errorf("expected WAVE format marker, got %s", string(wavData[8:12]))
	}
	if !bytes.Equal(wavData[12:16], []byte("fmt ")) {
		t.Errorf("expected 'fmt ' subchunk marker, got %s", string(wavData[12:16]))
	}
	if !bytes.Equal(wavData[36:40], []byte("data")) {
		t.Errorf("expected 'data' subchunk marker, got %s", string(wavData[36:40]))
	}
}

func TestSoundToggle(t *testing.T) {
	initial := Enabled
	defer func() {
		Enabled = initial
	}()

	toggled := Toggle()
	if toggled == initial || Enabled == initial {
		t.Errorf("expected Toggle to invert Enabled from %v to %v", initial, !initial)
	}

	toggled2 := Toggle()
	if toggled2 != initial || Enabled != initial {
		t.Errorf("expected second Toggle to restore Enabled to %v", initial)
	}
}

func TestPlayFunctionsWhenDisabled(t *testing.T) {
	initial := Enabled
	Enabled = false
	defer func() {
		Enabled = initial
	}()

	// Must be safe no-op without panics or blocking
	PlaySuccess()
	PlayError()
	PlayBreakpoint()
	PlayBell()
}

func TestInitSoundFiles(t *testing.T) {
	initSoundFiles()

	if successWavPath == "" || errorWavPath == "" || breakWavPath == "" || bellWavPath == "" {
		t.Fatalf("expected sound file paths to be set")
	}

	for _, path := range []string{successWavPath, errorWavPath, breakWavPath, bellWavPath} {
		fi, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected sound file to exist at %s: %v", path, err)
			continue
		}
		if fi.Size() < 44 {
			t.Errorf("sound file %s size too small: %d bytes", path, fi.Size())
		}
	}
}

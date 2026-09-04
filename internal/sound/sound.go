package sound

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	Enabled = true
	once    sync.Once
	tmpDir  string

	successWavPath string
	errorWavPath   string
	breakWavPath   string
	bellWavPath    string
)

type tone struct {
	freq       float64
	durationMs int
}

func initSoundFiles() {
	tmpDir = filepath.Join(os.TempDir(), "turborust_sounds")
	_ = os.MkdirAll(tmpDir, 0755)

	// 1. Success Tone: Warm 2-step ascending beep (740Hz -> 1108Hz, soft tone)
	successWavPath = filepath.Join(tmpDir, "success.wav")
	_ = os.WriteFile(successWavPath, generateWarmConeTone([]tone{
		{freq: 740, durationMs: 55},
		{freq: 1108, durationMs: 85},
	}), 0644)

	// 2. Error Tone: Warm heavy low-frequency speaker thump (196Hz)
	errorWavPath = filepath.Join(tmpDir, "error.wav")
	_ = os.WriteFile(errorWavPath, generateWarmConeTone([]tone{
		{freq: 196, durationMs: 180},
	}), 0644)

	// 3. Breakpoint Tone: Short clean analog speaker click (880Hz, 35ms)
	breakWavPath = filepath.Join(tmpDir, "break.wav")
	_ = os.WriteFile(breakWavPath, generateWarmConeTone([]tone{
		{freq: 880, durationMs: 35},
	}), 0644)

	// 4. Bell Tone: Mild 784Hz single beep (45ms)
	bellWavPath = filepath.Join(tmpDir, "bell.wav")
	_ = os.WriteFile(bellWavPath, generateWarmConeTone([]tone{
		{freq: 784, durationMs: 45},
	}), 0644)
}

// generateWarmConeTone simulates an authentic 2.5-inch IBM paper-cone PC speaker:
func generateWarmConeTone(tones []tone) []byte {
	const sampleRate = 22050
	var floatSamples []float64

	for _, t := range tones {
		n := int(float64(sampleRate) * float64(t.durationMs) / 1000.0)
		periodSamples := float64(sampleRate) / t.freq

		attackSamples := sampleRate * 8 / 1000  // 8ms attack
		decaySamples := sampleRate * 15 / 1000 // 15ms decay

		for i := 0; i < n; i++ {
			phase := math.Mod(float64(i), periodSamples) / periodSamples
			var raw float64
			if phase < 0.5 {
				raw = 1.0
			} else {
				raw = -1.0
			}

			// Apply envelope for paper cone physics
			env := 1.0
			if i < attackSamples {
				env = float64(i) / float64(attackSamples)
			} else if i > n-decaySamples {
				env = float64(n-i) / float64(decaySamples)
			}

			floatSamples = append(floatSamples, raw*env)
		}
	}

	alpha := 0.22
	filtered := 0.0
	audioData := make([]byte, len(floatSamples))

	for i, x := range floatSamples {
		filtered = filtered + alpha*(x-filtered)
		sampleVal := 128.0 + filtered*48.0
		if sampleVal > 255 {
			sampleVal = 255
		} else if sampleVal < 0 {
			sampleVal = 0
		}
		audioData[i] = byte(sampleVal)
	}

	// Build standard 44-byte WAV header
	buf := new(bytes.Buffer)
	numChannels := uint16(1)
	bitsPerSample := uint16(8)
	byteRate := uint32(sampleRate) * uint32(numChannels) * uint32(bitsPerSample/8)
	blockAlign := numChannels * (bitsPerSample / 8)
	dataLen := uint32(len(audioData))

	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, numChannels)
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buf, binary.LittleEndian, byteRate)
	_ = binary.Write(buf, binary.LittleEndian, blockAlign)
	_ = binary.Write(buf, binary.LittleEndian, bitsPerSample)
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, dataLen)
	buf.Write(audioData)

	return buf.Bytes()
}

func play(pathProvider func() string) {
	if !Enabled {
		return
	}
	once.Do(initSoundFiles)
	path := pathProvider()

	go func() {
		switch runtime.GOOS {
		case "darwin":
			if path != "" {
				_ = exec.Command("afplay", path).Run()
			}
		case "windows":
			if path != "" {
				psCmd := fmt.Sprintf(`(New-Object Media.SoundPlayer '%s').PlaySync()`, path)
				_ = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).Run()
			} else {
				fmt.Print("\a")
			}
		default:
			if path != "" {
				if err := exec.Command("aplay", "-q", path).Run(); err != nil {
					_ = exec.Command("paplay", path).Run()
				}
			} else {
				fmt.Print("\a")
			}
		}
	}()
}

// Toggle toggles sound on/off
func Toggle() bool {
	Enabled = !Enabled
	if Enabled {
		PlayBell()
	}
	return Enabled
}

// PlaySuccess plays the classic 2-step ascending PC speaker beep
func PlaySuccess() {
	play(func() string { return successWavPath })
}

// PlayError plays the heavy low-frequency PC speaker buzz
func PlayError() {
	play(func() string { return errorWavPath })
}

// PlayBreakpoint plays the crisp piezo click/beep on breakpoint stop
func PlayBreakpoint() {
	play(func() string { return breakWavPath })
}

// PlayBell plays a single standard PC speaker beep
func PlayBell() {
	play(func() string { return bellWavPath })
}

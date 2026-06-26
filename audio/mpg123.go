//go:build nofakecgo

package audio

import (
	"fmt"
	"sync"

	"github.com/ebitengine/purego"
)

const (
	mpg123Done        int32 = -12
	mpg123NewFormat   int32 = -11
	mpg123NeedMore    int32 = -10
	mpg123OK          int32 = 0
	mpg123AddFlags    int32 = 2
	mpg123Quiet       int64 = 0x20
	mpg123Stereo      int32 = 2
	mpg123EncSigned16 int32 = 0x0d0
)

var (
	mpg123LoadOnce sync.Once
	mpg123LoadErr  error

	mpg123Init           func() int32
	mpg123Exit           func()
	mpg123New            func(decoder string, errcode *int32) uintptr
	mpg123Delete         func(mh uintptr)
	mpg123OpenFeed       func(mh uintptr) int32
	mpg123Close          func(mh uintptr) int32
	mpg123Feed           func(mh uintptr, input []byte, size uintptr) int32
	mpg123Read           func(mh uintptr, output []byte, size uintptr, done *uintptr) int32
	mpg123GetFormat      func(mh uintptr, rate *int64, channels *int32, encoding *int32) int32
	mpg123FormatNone     func(mh uintptr) int32
	mpg123Format         func(mh uintptr, rate int64, channels int32, encodings int32) int32
	mpg123Param          func(mh uintptr, parameter int32, value int64, fvalue float64) int32
	mpg123PlainStrerror  func(errcode int32) string
	mpg123HandleStrerror func(mh uintptr) string
)

func decodeMPG123PCM(data []byte) ([]byte, int, string, error) {
	if len(data) == 0 {
		return nil, 0, "", fmt.Errorf("empty mp3 data")
	}
	if err := loadMPG123(); err != nil {
		return nil, 0, "", err
	}
	var newErr int32
	handle := mpg123New("", &newErr)
	if handle == 0 {
		return nil, 0, "", fmt.Errorf("mpg123_new: %s", mpg123Error(newErr))
	}
	defer mpg123Delete(handle)
	if err := mpg123Check(handle, mpg123Param(handle, mpg123AddFlags, mpg123Quiet, 0), "mpg123_param"); err != nil {
		return nil, 0, "", err
	}

	if err := mpg123Check(handle, mpg123FormatNone(handle), "mpg123_format_none"); err != nil {
		return nil, 0, "", err
	}
	for _, rate := range []int64{8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000} {
		if err := mpg123Check(handle, mpg123Format(handle, rate, mpg123Stereo, mpg123EncSigned16), "mpg123_format"); err != nil {
			return nil, 0, "", err
		}
	}
	if err := mpg123Check(handle, mpg123OpenFeed(handle), "mpg123_open_feed"); err != nil {
		return nil, 0, "", err
	}
	defer mpg123Close(handle)

	if code := mpg123Feed(handle, data, uintptr(len(data))); code != mpg123OK && code != mpg123NeedMore {
		return nil, 0, "", fmt.Errorf("mpg123_feed: %s", mpg123HandleError(handle, code))
	}

	buffer := make([]byte, 64*1024)
	var pcm []byte
	var sampleRate int
	for {
		var done uintptr
		code := mpg123Read(handle, buffer, uintptr(len(buffer)), &done)
		if done > 0 {
			pcm = append(pcm, buffer[:done]...)
		}
		switch code {
		case mpg123OK:
			if sampleRate == 0 {
				rate, err := mpg123CurrentFormat(handle)
				if err != nil {
					return nil, 0, "", err
				}
				sampleRate = rate
			}
		case mpg123NewFormat:
			rate, err := mpg123CurrentFormat(handle)
			if err != nil {
				return nil, 0, "", err
			}
			if sampleRate != 0 && sampleRate != rate {
				return nil, 0, "", fmt.Errorf("mpg123 format changed %d -> %d", sampleRate, rate)
			}
			sampleRate = rate
		case mpg123NeedMore, mpg123Done:
			if len(pcm)%4 != 0 {
				return nil, 0, "", fmt.Errorf("invalid mpg123 stereo pcm length %d", len(pcm))
			}
			if sampleRate == 0 {
				rate, err := mpg123CurrentFormat(handle)
				if err != nil {
					return nil, 0, "", err
				}
				sampleRate = rate
			}
			return pcm, sampleRate, "mpg123", nil
		default:
			return nil, 0, "", fmt.Errorf("mpg123_read: %s", mpg123HandleError(handle, code))
		}
	}
}

func loadMPG123() error {
	mpg123LoadOnce.Do(func() {
		handle, err := openDynamicLibrary([]string{
			"libmpg123.so.0",
			"libmpg123.so",
			"libmpg123.dylib",
			"mpg123.dll",
			"libmpg123-0.dll",
		})
		if handle == 0 {
			mpg123LoadErr = fmt.Errorf("load mpg123: %w", err)
			return
		}

		purego.RegisterLibFunc(&mpg123Init, handle, "mpg123_init")
		purego.RegisterLibFunc(&mpg123Exit, handle, "mpg123_exit")
		purego.RegisterLibFunc(&mpg123New, handle, "mpg123_new")
		purego.RegisterLibFunc(&mpg123Delete, handle, "mpg123_delete")
		purego.RegisterLibFunc(&mpg123OpenFeed, handle, "mpg123_open_feed")
		purego.RegisterLibFunc(&mpg123Close, handle, "mpg123_close")
		purego.RegisterLibFunc(&mpg123Feed, handle, "mpg123_feed")
		purego.RegisterLibFunc(&mpg123Read, handle, "mpg123_read")
		purego.RegisterLibFunc(&mpg123GetFormat, handle, "mpg123_getformat")
		purego.RegisterLibFunc(&mpg123FormatNone, handle, "mpg123_format_none")
		purego.RegisterLibFunc(&mpg123Format, handle, "mpg123_format")
		purego.RegisterLibFunc(&mpg123Param, handle, "mpg123_param2")
		purego.RegisterLibFunc(&mpg123PlainStrerror, handle, "mpg123_plain_strerror")
		purego.RegisterLibFunc(&mpg123HandleStrerror, handle, "mpg123_strerror")

		if code := mpg123Init(); code != mpg123OK {
			mpg123LoadErr = fmt.Errorf("mpg123_init: %s", mpg123Error(code))
			return
		}
	})
	return mpg123LoadErr
}

func mpg123CurrentFormat(handle uintptr) (int, error) {
	var rate int64
	var channels int32
	var encoding int32
	if err := mpg123Check(handle, mpg123GetFormat(handle, &rate, &channels, &encoding), "mpg123_getformat"); err != nil {
		return 0, err
	}
	if rate <= 0 {
		return 0, fmt.Errorf("mpg123 invalid sample rate %d", rate)
	}
	if channels != 2 || encoding != mpg123EncSigned16 {
		return 0, fmt.Errorf("mpg123 unsupported output format rate=%d channels=%d encoding=0x%X", rate, channels, encoding)
	}
	return int(rate), nil
}

func mpg123Check(handle uintptr, code int32, operation string) error {
	if code == mpg123OK {
		return nil
	}
	return fmt.Errorf("%s: %s", operation, mpg123HandleError(handle, code))
}

func mpg123HandleError(handle uintptr, code int32) string {
	if handle != 0 && mpg123HandleStrerror != nil {
		if msg := mpg123HandleStrerror(handle); msg != "" {
			return fmt.Sprintf("%s (%d)", msg, code)
		}
	}
	return mpg123Error(code)
}

func mpg123Error(code int32) string {
	if mpg123PlainStrerror != nil {
		if msg := mpg123PlainStrerror(code); msg != "" {
			return fmt.Sprintf("%s (%d)", msg, code)
		}
	}
	return fmt.Sprintf("error %d", code)
}

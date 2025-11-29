module github.com/user/audio-dubbing-system

go 1.21

require (
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-20240101000000-000000000000
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

// Use local whisper.cpp if available
replace github.com/ggerganov/whisper.cpp/bindings/go => ./third_party/whisper.cpp/bindings/go

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func isNameMatching(filenameWithExtension string, audioAttr string) bool {
	extension := filepath.Ext(filenameWithExtension)
	filename := strings.Split(filenameWithExtension, extension)[0]
	return filename == audioAttr
}

func getAudioFilePath(audioName *string) string {
	var audioFilePath string = ""
	usr, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	path := filepath.Join(usr, ".sb/audios")
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filename := info.Name()
			if isNameMatching(filename, *audioName) {
				audioFilePath = path
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if audioFilePath == "" {
		log.Fatal("No audio found named ", *audioName)
	}
	return audioFilePath
}

func decode(file *os.File) (s beep.StreamSeekCloser, format beep.Format, err error) {
	extension := filepath.Ext(file.Name())
	if extension == ".mp3" {
		return mp3.Decode(file)
	} else if extension == ".wav" {
		return wav.Decode(file)
	} else {
		return nil, beep.Format{}, errors.New("invalid file extension: " + extension)
	}
}

func playAudio(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))
	<-done
}

var audioName string

const (
	defaultValue = ""
	usage        = "the name of the audio to play"
)

func main() {
	flag.StringVar(&audioName, "audio", defaultValue, usage)
	flag.StringVar(&audioName, "a", defaultValue, usage)
	flag.Parse()

	if audioName == "" {
		log.Fatal("specify an audio name")
	}

	audioFilePath := getAudioFilePath(&audioName)
	playAudio(audioFilePath)
}

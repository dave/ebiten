// Copyright 2017 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"

	"github.com/dave/ebiten"
	"github.com/dave/ebiten/audio"
	"github.com/dave/ebiten/audio/wav"
	"github.com/dave/ebiten/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240
	sampleRate   = 44100
)

var (
	audioContext *audio.Context
	audioPlayer  *audio.Player
)

func init() {
	var err error
	// Initialize audio context.
	audioContext, err = audio.NewContext(sampleRate)
	if err != nil {
		log.Fatal(err)
	}

	// Open a wav file.
	// Note that f.Close() should not be closed in this init function
	// since audio.Player manages stream state.
	f, err := ebitenutil.OpenFile("_resources/audio/jab.wav")
	if err != nil {
		log.Fatal(err)
	}

	// Decode wav-formatted data and retrieve decoded PCM stream.
	d, err := wav.Decode(audioContext, f)
	if err != nil {
		log.Fatal(err)
	}

	// Create an audio.Player that has one stream.
	audioPlayer, err = audio.NewPlayer(audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
}

func update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyP) && !audioPlayer.IsPlaying() {
		// As audioPlayer has one stream and remembers the playing position,
		// rewinding is needed before playing when reusing audioPlayer.
		audioPlayer.Rewind()
		audioPlayer.Play()
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	if audioPlayer.IsPlaying() {
		ebitenutil.DebugPrint(screen, "Bump!")
	} else {
		ebitenutil.DebugPrint(screen, "Press P to play the wav")
	}
	return nil
}

func main() {
	if err := ebiten.Run(update, screenWidth, screenHeight, 2, "WAV (Ebiten Demo)"); err != nil {
		log.Fatal(err)
	}
}

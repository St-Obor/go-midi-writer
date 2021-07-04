package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gitlab.com/gomidi/midi/writer"
)

var numberOfTracks = 8

type Note struct {
	Track        uint8
	MidiNote     uint8
	BeatDuration uint8
	BeatsPlayed  uint8
	Ringing      bool
	Strike       bool
}

func inSlice(needle int, haystack []int) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func main() {

	// Setup mode
	allowedNotes := []int{
		49, 51, 54, 56, 58,
		61, 63, 66, 68, 70,
		73, 75, 78, 80, 82,
	}

	// Setup every possible note
	notes := make([]Note, 120) // Midi range 0 - 120
	for i := range notes {
		notes[i].BeatDuration = 1
	}

	// Setup matrix
	matrix := make([][][]uint32, numberOfTracks)
	for track := range matrix {
		lengths := []uint32{8, 16, 24, 32, 64, 128}
		matrix[track] = make([][]uint32, lengths[rand.Intn(len(lengths))])
		for beat := range matrix[track] {
			matrix[track][beat] = make([]uint32, 120)
		}
	}

	// Create some data
	min := 0
	max := 100
	for track := range matrix {
		for beat := range matrix[track] {
			for midiNote := range matrix[track][beat] {
				if inSlice(midiNote, allowedNotes) {
					if (rand.Intn(max-min+1) + min) >= 98 {
						matrix[track][beat][midiNote] = 1 // Set note to "on"
					}
				}
			}
		}
	}

	//
	defaultFilename := "session-" + strconv.Itoa(int(time.Now().Unix())) + "-test.mid"
	filePath := "./" + defaultFilename

	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
	}()

	err = writer.WriteSMF(filePath, uint16(numberOfTracks), func(wr *writer.SMF) error {

		//
		if len(notes) != len(matrix[0][0]) {
			log.Fatal("Number of playable notes and number of possible notes at each beat must be equal.")
		}

		// For each track/channel
		for track := range matrix {

			//
			fmt.Println("Writing track/channel:", track)

			// Set track name
			writer.TrackSequenceName(wr, "Track: "+strconv.Itoa(track))

			// Set channel
			wr.SetChannel(uint8(track))

			// For each beat (x axis) in this track
			for beat := range matrix[track] {

				// For each note position (y axis) at this beat
				for i := range matrix[track][beat] {
					// If there are any "on" events, set Strike to true
					if matrix[track][beat][i] == 1 {
						notes[i].Strike = true
					}
				}

				// For each Note in notes
				for i := range notes {
					// If this Note is Ringing increment its BeatsPlayed
					if notes[i].Ringing {
						notes[i].BeatsPlayed++
						// If BeatsPlayed is >= BeatDuraction
						if notes[i].BeatsPlayed >= notes[i].BeatDuration {
							// Turn Note off
							writer.NoteOff(wr, uint8(i))
						}
					}

					// If note is set to be Struck
					if notes[i].Strike {
						// If this Note is already Ringing, turn it off
						if notes[i].Ringing {
							// Turn Note off
							writer.NoteOff(wr, uint8(i))
						}

						// Turn this Note on
						writer.NoteOn(wr, uint8(i), 100)

						// Update Note state
						notes[i].BeatsPlayed = 0
						notes[i].Strike = false
						notes[i].Ringing = true
					}
				}

				// Move forward by 1 beat
				writer.Forward(wr, 0, 1, 4)
			}

			// End this track/channel
			writer.EndOfTrack(wr)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("could not write SMF file %v\n", f)
		return
	}
}

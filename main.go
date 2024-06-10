package main

import dumplingai "github.com/nt54hamnghi/hoc/dumpling-ai"

func main() {
	transcript, err := dumplingai.GetTranscript("https://www.youtube.com/watch?v=olBvyjMuqIw")
	if err != nil {
		panic(err)
	}
	println(transcript)
}

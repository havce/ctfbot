package discord

import "math/rand"

func cheer() string {
	cheers := []string{
		"Hooray",
		"Woo-hoo",
		"Cheers",
		"Yippee",
		"Yay",
		"Let's go",
		"Hip, hip, hooray",
		"Fantastic",
		"Celebrate",
		"Party time",
	}

	return cheers[rand.Intn(len(cheers))]
}

package discord

import "slices"

var blocklist []string = []string{}

func (s *Server) canIFlagHere(name string) bool {
	if name == s.GeneralChannel || name == s.RegistrationChannel {
		return false
	}

	// If not found in the blocklist.
	return !slices.Contains(blocklist, name)
}

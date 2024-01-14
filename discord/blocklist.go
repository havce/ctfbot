package discord

import "slices"

var blocklist []string = []string{}

// flagAllowed is a helper function used by handleFlag that
// checks if name clashes with the General and Registration channel and
// an additional blocklist.
func (s *Server) flagAllowed(name string) bool {
	if name == s.GeneralChannel || name == s.RegistrationChannel {
		return false
	}

	// If not found in the blocklist.
	return !slices.Contains(blocklist, name)
}

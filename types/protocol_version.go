package types

type ProtocolId = uint16

type ProtocolVersion struct {
	ID    ProtocolId
	Names []string
}

var ProtocolVersions = struct {
	ZERO              ProtocolVersion
	MINECRAFT_1_18    ProtocolVersion
	MINECRAFT_1_18_2  ProtocolVersion
	MINECRAFT_1_19    ProtocolVersion
	MINECRAFT_1_19_1  ProtocolVersion
	MINECRAFT_1_19_3  ProtocolVersion
	MINECRAFT_1_19_4  ProtocolVersion
	MINECRAFT_1_20    ProtocolVersion
	MINECRAFT_1_20_2  ProtocolVersion
	MINECRAFT_1_20_3  ProtocolVersion
	MINECRAFT_1_20_5  ProtocolVersion
	MINECRAFT_1_21    ProtocolVersion
	MINECRAFT_1_21_2  ProtocolVersion
	MINECRAFT_1_21_4  ProtocolVersion
	MINECRAFT_1_21_5  ProtocolVersion
	MINECRAFT_1_21_6  ProtocolVersion
	MINECRAFT_1_21_7  ProtocolVersion
	MINECRAFT_1_21_9  ProtocolVersion
	MINECRAFT_1_21_11 ProtocolVersion
	MINECRAFT_26_1    ProtocolVersion
	MINECRAFT_26_2    ProtocolVersion
}{
	ZERO: ProtocolVersion{ID: 0, Names: []string{}},

	// 1.18.1 stayed on 1.18's protocol, so a client on either of them is a
	// client on this version. It is the oldest this server speaks, and like
	// the six above it from before the configuration phase: see
	// HasConfigurationPhase. It is from before the profile key as well, so a
	// client on it always encrypts the encryption challenge: see
	// MaySignEncryptionChallenge.
	MINECRAFT_1_18: ProtocolVersion{ID: 757, Names: []string{"1.18", "1.18.1"}},

	// 1.18.2 has 758 to itself: 1.18 and 1.18.1 share 757 below it and 1.19
	// moved to 759. Like the five above it, it is from before the
	// configuration phase: see HasConfigurationPhase. It is from before the
	// profile key as well, so a client on it always encrypts the encryption
	// challenge: see MaySignEncryptionChallenge.
	MINECRAFT_1_18_2: ProtocolVersion{ID: 758, Names: []string{"1.18.2"}},

	// 1.19 has 759 to itself: 1.18.2 sits on 758 and 1.19.1 moved to 760.
	// Like the four above it, it is from before the configuration phase: see
	// HasConfigurationPhase. It is the first on which a client may sign the
	// encryption challenge rather than encrypt it, since 1.19 is where the
	// profile key appeared: see MaySignEncryptionChallenge.
	MINECRAFT_1_19: ProtocolVersion{ID: 759, Names: []string{"1.19"}},

	// 1.19.2 stayed on 1.19.1's protocol, so a client on either of them is a
	// client on this version. Like the three above it, it is from before the
	// configuration phase: see HasConfigurationPhase. It is also the last on
	// which a client may sign the encryption challenge rather than encrypt
	// it: see MaySignEncryptionChallenge.
	MINECRAFT_1_19_1: ProtocolVersion{ID: 760, Names: []string{"1.19.1", "1.19.2"}},

	// 1.19.3 has 761 to itself: 1.19.2 sits on 760 and 1.19.4 moved to 762.
	// It is from before the configuration phase as well.
	MINECRAFT_1_19_3: ProtocolVersion{ID: 761, Names: []string{"1.19.3"}},

	// 1.19.4 has 762 to itself: 1.19.3 sits on 761 and 1.20 moved to 763. It
	// is from before the configuration phase as well.
	MINECRAFT_1_19_4: ProtocolVersion{ID: 762, Names: []string{"1.19.4"}},

	// 1.20.1 stayed on 1.20's protocol, so a client on either of them is a
	// client on this version. It is the last before the configuration phase:
	// see HasConfigurationPhase.
	MINECRAFT_1_20: ProtocolVersion{ID: 763, Names: []string{"1.20", "1.20.1"}},

	// 1.20.2 has 764 to itself: 1.20 and 1.20.1 share 763 below it, and 1.20.3
	// moved to 765.
	MINECRAFT_1_20_2: ProtocolVersion{ID: 764, Names: []string{"1.20.2"}},

	// 1.20.4 stayed on 1.20.3's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_20_3: ProtocolVersion{ID: 765, Names: []string{"1.20.3", "1.20.4"}},

	// 1.20.6 stayed on 1.20.5's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_20_5: ProtocolVersion{ID: 766, Names: []string{"1.20.5", "1.20.6"}},

	// 1.21.1 stayed on 1.21's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21: ProtocolVersion{ID: 767, Names: []string{"1.21", "1.21.1"}},

	// 1.21.3 stayed on 1.21.2's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21_2: ProtocolVersion{ID: 768, Names: []string{"1.21.2", "1.21.3"}},

	// 1.21.4 has 769 to itself: 1.21.2 and 1.21.3 share 768 below it, and
	// 1.21.5 moved to 770.
	MINECRAFT_1_21_4: ProtocolVersion{ID: 769, Names: []string{"1.21.4"}},

	// 1.21.5 has 770 to itself: 1.21.4 sits on 769 and 1.21.6 moved to 771.
	MINECRAFT_1_21_5: ProtocolVersion{ID: 770, Names: []string{"1.21.5"}},

	// 1.21.6 has 771 to itself: 1.21.5 sits on 770 and 1.21.7 moved to 772.
	MINECRAFT_1_21_6: ProtocolVersion{ID: 771, Names: []string{"1.21.6"}},

	// 1.21.8 stayed on 1.21.7's protocol, so a client on either of them is a
	// client on this version.
	MINECRAFT_1_21_7: ProtocolVersion{ID: 772, Names: []string{"1.21.7", "1.21.8"}},

	// 1.21.10 stayed on 1.21.9's protocol, so a client on either of them is a
	// client on this version; 1.21.11 has 774 to itself.
	MINECRAFT_1_21_9:  ProtocolVersion{ID: 773, Names: []string{"1.21.9", "1.21.10"}},
	MINECRAFT_1_21_11: ProtocolVersion{ID: 774, Names: []string{"1.21.11"}},

	// The three releases of the 26.1 cycle share a protocol, so a client on any
	// of them is a client on this version.
	MINECRAFT_26_1: ProtocolVersion{ID: 775, Names: []string{"26.1", "26.1.1", "26.1.2"}},
	MINECRAFT_26_2: ProtocolVersion{ID: 776, Names: []string{"26.2"}},
}

// SupportedProtocolVersions is every version a client may connect on, oldest
// first.
//
// The order is the one the packet transformers walk. A packet read from a
// client is carried up this list one version at a time until it reaches the
// latest, which is the only version anything is implemented at, and a packet
// written to a client is carried back down it. That is why the list has to hold
// every version in between rather than only the ends: a step is what a
// transformer is registered for.
//
// ZERO is not among them. It is what a connection speaks before its handshake
// says otherwise, which is not a version anything is transformed to or from.
var SupportedProtocolVersions = []ProtocolVersion{
	ProtocolVersions.MINECRAFT_1_18,
	ProtocolVersions.MINECRAFT_1_18_2,
	ProtocolVersions.MINECRAFT_1_19,
	ProtocolVersions.MINECRAFT_1_19_1,
	ProtocolVersions.MINECRAFT_1_19_3,
	ProtocolVersions.MINECRAFT_1_19_4,
	ProtocolVersions.MINECRAFT_1_20,
	ProtocolVersions.MINECRAFT_1_20_2,
	ProtocolVersions.MINECRAFT_1_20_3,
	ProtocolVersions.MINECRAFT_1_20_5,
	ProtocolVersions.MINECRAFT_1_21,
	ProtocolVersions.MINECRAFT_1_21_2,
	ProtocolVersions.MINECRAFT_1_21_4,
	ProtocolVersions.MINECRAFT_1_21_5,
	ProtocolVersions.MINECRAFT_1_21_6,
	ProtocolVersions.MINECRAFT_1_21_7,
	ProtocolVersions.MINECRAFT_1_21_9,
	ProtocolVersions.MINECRAFT_1_21_11,
	ProtocolVersions.MINECRAFT_26_1,
	ProtocolVersions.MINECRAFT_26_2,
}

// LatestProtocolVersion is the version every packet is implemented at. Older
// versions are reached by transforming what this one produces, so adding one
// adds a table of ids and whatever transformers its differences need, and no
// second implementation of anything.
var LatestProtocolVersion = SupportedProtocolVersions[len(SupportedProtocolVersions)-1]

var protocolVersionsById = map[ProtocolId]ProtocolVersion{
	ProtocolVersions.ZERO.ID:              ProtocolVersions.ZERO,
	ProtocolVersions.MINECRAFT_1_18.ID:    ProtocolVersions.MINECRAFT_1_18,
	ProtocolVersions.MINECRAFT_1_18_2.ID:  ProtocolVersions.MINECRAFT_1_18_2,
	ProtocolVersions.MINECRAFT_1_19.ID:    ProtocolVersions.MINECRAFT_1_19,
	ProtocolVersions.MINECRAFT_1_19_1.ID:  ProtocolVersions.MINECRAFT_1_19_1,
	ProtocolVersions.MINECRAFT_1_19_3.ID:  ProtocolVersions.MINECRAFT_1_19_3,
	ProtocolVersions.MINECRAFT_1_19_4.ID:  ProtocolVersions.MINECRAFT_1_19_4,
	ProtocolVersions.MINECRAFT_1_20.ID:    ProtocolVersions.MINECRAFT_1_20,
	ProtocolVersions.MINECRAFT_1_20_2.ID:  ProtocolVersions.MINECRAFT_1_20_2,
	ProtocolVersions.MINECRAFT_1_20_3.ID:  ProtocolVersions.MINECRAFT_1_20_3,
	ProtocolVersions.MINECRAFT_1_20_5.ID:  ProtocolVersions.MINECRAFT_1_20_5,
	ProtocolVersions.MINECRAFT_1_21.ID:    ProtocolVersions.MINECRAFT_1_21,
	ProtocolVersions.MINECRAFT_1_21_2.ID:  ProtocolVersions.MINECRAFT_1_21_2,
	ProtocolVersions.MINECRAFT_1_21_4.ID:  ProtocolVersions.MINECRAFT_1_21_4,
	ProtocolVersions.MINECRAFT_1_21_5.ID:  ProtocolVersions.MINECRAFT_1_21_5,
	ProtocolVersions.MINECRAFT_1_21_6.ID:  ProtocolVersions.MINECRAFT_1_21_6,
	ProtocolVersions.MINECRAFT_1_21_7.ID:  ProtocolVersions.MINECRAFT_1_21_7,
	ProtocolVersions.MINECRAFT_1_21_9.ID:  ProtocolVersions.MINECRAFT_1_21_9,
	ProtocolVersions.MINECRAFT_1_21_11.ID: ProtocolVersions.MINECRAFT_1_21_11,
	ProtocolVersions.MINECRAFT_26_1.ID:    ProtocolVersions.MINECRAFT_26_1,
	ProtocolVersions.MINECRAFT_26_2.ID:    ProtocolVersions.MINECRAFT_26_2,
}

func GetProtocolVersionById(id ProtocolId) ProtocolVersion {
	if v, ok := protocolVersionsById[id]; ok {
		return v
	}

	return ProtocolVersions.ZERO
}

// protocolVersionIndex is where version sits in SupportedProtocolVersions, or
// -1 for a version that is not on the chain at all.
func protocolVersionIndex(version ProtocolVersion) int {
	for i, supported := range SupportedProtocolVersions {
		if supported.ID == version.ID {
			return i
		}
	}

	return -1
}

// IsSupportedProtocolVersion reports whether version is one the server speaks,
// which is to say one the transformers can carry a packet to and from.
func IsSupportedProtocolVersion(version ProtocolVersion) bool {
	return protocolVersionIndex(version) >= 0
}

// NextProtocolVersion returns the version one step newer than version. It
// reports false for the latest version, which has nothing above it, and for a
// version that is not on the chain.
func NextProtocolVersion(version ProtocolVersion) (ProtocolVersion, bool) {
	index := protocolVersionIndex(version)
	if index < 0 || index == len(SupportedProtocolVersions)-1 {
		return ProtocolVersions.ZERO, false
	}

	return SupportedProtocolVersions[index+1], true
}

// PreviousProtocolVersion returns the version one step older than version. It
// reports false for the oldest version and for a version that is not on the
// chain.
func PreviousProtocolVersion(version ProtocolVersion) (ProtocolVersion, bool) {
	index := protocolVersionIndex(version)
	if index <= 0 {
		return ProtocolVersions.ZERO, false
	}

	return SupportedProtocolVersions[index-1], true
}

// HasConfigurationPhase reports whether a client on this version passes
// through the configuration phase on its way from the login to the play
// phase. 1.20.2 is where the phase appeared. A client before it -- 1.20,
// 1.19.4, 1.19.3, 1.19.1, 1.19, 1.18.2 and 1.18 -- is in play the moment
// its login succeeds, with nothing acknowledged in between, and what the phase carries
// from 1.20.2 on -- the registries and the tags -- reaches such a client
// through the play phase instead: the registries inside the play login
// packet itself, and the tags as a play packet right after it.
func (v ProtocolVersion) HasConfigurationPhase() bool {
	return v.ID >= ProtocolVersions.MINECRAFT_1_20_2.ID
}

// MaySignEncryptionChallenge reports whether a client on this version may
// answer the encryption request's challenge with a signature instead of
// encrypting it. 1.19 gave the client a profile key, signed by Mojang, and a
// client holding one -- which is every client logged into an account --
// signs the challenge under it and never encrypts it; a client without one
// encrypts it as every version does. 1.19.3 is where the key left the login,
// so from it on the challenge is always encrypted, and 1.18.2 and 1.18,
// from before the key, have nothing to sign with and encrypt it as well:
// 1.19 and 1.19.1 are the two versions that may sign. A signature is a thing this server
// cannot check, since it does not keep the key the client sent with its
// hello, and a version that may sign is a version whose response is let
// through without the challenge: see CompleteEncryption.
func (v ProtocolVersion) MaySignEncryptionChallenge() bool {
	return IsSupportedProtocolVersion(v) && v.ID >= ProtocolVersions.MINECRAFT_1_19.ID && v.ID < ProtocolVersions.MINECRAFT_1_19_3.ID
}

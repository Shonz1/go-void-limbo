package types

type ProtocolId = uint16

type ProtocolVersion struct {
	ID    ProtocolId
	Names []string
}

var ProtocolVersions = struct {
	ZERO           ProtocolVersion
	MINECRAFT_26_2 ProtocolVersion
}{
	ZERO:           ProtocolVersion{ID: 0, Names: []string{}},
	MINECRAFT_26_2: ProtocolVersion{ID: 776, Names: []string{"26.2"}},
}

var protocolVersionsById = map[ProtocolId]ProtocolVersion{
	ProtocolVersions.ZERO.ID:           ProtocolVersions.ZERO,
	ProtocolVersions.MINECRAFT_26_2.ID: ProtocolVersions.MINECRAFT_26_2,
}

func GetProtocolVersionById(id ProtocolId) ProtocolVersion {
	if v, ok := protocolVersionsById[id]; ok {
		return v
	}

	return ProtocolVersions.ZERO
}

package types

import (
	"reflect"
)

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

func GetProtocolVersionById(id ProtocolId) ProtocolVersion {
	v := reflect.ValueOf(ProtocolVersions)

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		typedValue, ok := value.Interface().(ProtocolVersion)
		if !ok {
			continue
		}

		if typedValue.ID == id {
			return typedValue
		}
	}

	return ProtocolVersions.ZERO
}

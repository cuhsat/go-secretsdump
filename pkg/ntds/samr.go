package ntds

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type SAMRUserProperties struct {
	Reserved1         uint32
	Length            uint32
	Reserved2         uint16
	Reserved3         uint16
	Reserved4         [96]byte
	PropertySignature uint16
	PropertyCount     uint16
	Properties        []SAMRUserProperty
}

type SAMRKerbStoredCredNew struct {
	Revision, Flags, CredentialCount,
	ServiceCredentialCount, OldCredentialCount,
	OlderCredentialCount, DefaultSaltLength,
	DefaultSaltMaximumLength uint16
	DefaultSaltOffset, DefaultIterationCount uint32
	Buffer                                   []byte
}

func NewSAMRKerbStoredCredNew(d []byte) SAMRKerbStoredCredNew {
	r := SAMRKerbStoredCredNew{}
	curs := 0
	r.Revision = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.Flags = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.CredentialCount = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.ServiceCredentialCount = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.OldCredentialCount = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.OlderCredentialCount = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.DefaultSaltLength = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.DefaultSaltMaximumLength = binary.LittleEndian.Uint16(getAndMoveCursor(d, &curs, 2))
	r.DefaultSaltOffset = binary.LittleEndian.Uint32(getAndMoveCursor(d, &curs, 4))
	r.DefaultIterationCount = binary.LittleEndian.Uint32(getAndMoveCursor(d, &curs, 4))
	r.Buffer = d[curs:]

	return r
}

type SAMRKerbKeyDataNew struct {
	Reserved1, Reserved2 uint16
	Reserved3, IterationCount,
	KeyType, KeyLength, KeyOffset uint32
}

func NewSAMRKerbKeyDataNew(d []byte) SAMRKerbKeyDataNew {
	kd := SAMRKerbKeyDataNew{}
	_ = binary.Read(bytes.NewReader(d), binary.LittleEndian, &kd)
	return kd
}

func getAndMoveCursor(data []byte, curs *int, size int) []byte {
	d := data[*curs : *curs+size]
	*curs += size
	return d
}

func NewSAMRUserProperties(data []byte) SAMRUserProperties {
	r := SAMRUserProperties{}
	curs := 0

	r.Reserved1 = binary.LittleEndian.Uint32(getAndMoveCursor(data, &curs, 4))
	r.Length = binary.LittleEndian.Uint32(getAndMoveCursor(data, &curs, 4))
	r.Reserved2 = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
	r.Reserved3 = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
	copy(r.Reserved4[:], data[curs:curs+96])
	curs += 96
	r.PropertySignature = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
	if len(data) > curs+2 {
		r.PropertyCount = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
		//fill properties
		for i := uint16(0); i < r.PropertyCount; i++ {
			np := SAMRUserProperty{}
			np.NameLength = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
			np.ValueLength = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
			np.Reserved = binary.LittleEndian.Uint16(getAndMoveCursor(data, &curs, 2))
			np.PropertyName = data[curs : curs+int(np.NameLength)]
			curs += int(np.NameLength)
			np.PropertyValue = data[curs : curs+int(np.ValueLength)]
			curs += int(np.ValueLength)
			r.Properties = append(r.Properties, np)
		}
	}
	return r
}

type SAMRUserProperty struct {
	NameLength    uint16
	ValueLength   uint16
	Reserved      uint16
	PropertyName  []byte
	PropertyValue []byte
}

type SAMRRPCSID struct {
	Revision            uint8   //'<B'
	SubAuthorityCount   uint8   //'<B'
	IdentifierAuthority [6]byte //SAMR_RPC_SID_IDENTIFIER_AUTHORITY
	SubLen              int     //    ('SubLen','_-SubAuthority','self["SubAuthorityCount"]*4'),
	SubAuthority        []byte  //':'
}

func (s SAMRRPCSID) Rid() uint32 {
	l := s.SubAuthorityCount
	return binary.BigEndian.Uint32(s.SubAuthority[(l-1)*4 : (l-1)*4+4])
}

func NewSAMRRPCSID(data []byte) (SAMRRPCSID, error) {
	r := SAMRRPCSID{}
	if len(data) < 6 {
		return r, fmt.Errorf("bad SAMR data: %s", string(data))
	}
	curs := 0

	r.Revision = data[0]
	r.SubAuthorityCount = data[1]

	getAndMoveCursor(data, &curs, 2)
	copy(r.IdentifierAuthority[:], getAndMoveCursor(data, &curs, 6))
	r.SubLen = int(r.SubAuthorityCount) * 4
	r.SubAuthority = data[curs:] // make([]byte, len(data[curs:]))
	//copy(r.SubAuthority, data[curs:])
	return r, nil
}

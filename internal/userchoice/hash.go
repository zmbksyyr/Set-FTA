package userchoice

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

const experience = "User Choice set via Windows User Experience {D18B6DD5-6124-4341-9318-804003BAFA0B}"

// Hash returns the Windows UserChoice validation hash for one association.
func Hash(extension, sid, progID, timestamp string) string {
	base := strings.ToLower(extension + sid + progID + timestamp + experience)
	encoded := utf16.Encode([]rune(base))
	encoded = append(encoded, 0)
	data := make([]byte, len(encoded)*2)
	for i, value := range encoded {
		binary.LittleEndian.PutUint16(data[i*2:], value)
	}

	digest := md5.Sum(data)
	lengthBase := len(base)*2 + 2
	length := boolInt((lengthBase&4) <= 1) + (lengthBase >> 2) - 1
	if length <= 1 {
		return ""
	}

	var cache, out1, out2 uint32
	position := 0
	counter := ((length - 2) >> 1) + 1
	md51 := (word(digest[:], 0) | 1) + 0x69FB0000
	md52 := (word(digest[:], 4) | 1) + 0x13DB0000

	for ; counter > 0; counter-- {
		r0 := word(data, position) + out1
		r1 := word(data, position+4)
		position += 8
		r2 := r0*md51 - 0x10FA9605*(r0>>16)
		r2 = 0x79F8A395*r2 + 0x689B6B9F*(r2>>16)
		r3 := 0xEA970001*r2 - 0x3C101569*(r2>>16)
		r4 := r3 + r1
		r5 := cache + r3
		r6 := r4*md52 - 0x3CE8EC25*(r4>>16)
		r6 = 0x59C3AF2D*r6 - 0x2232E0F1*(r6>>16)
		out1 = 0x1EC90001*r6 + 0x35BD1EC9*(r6>>16)
		out2 = r5 + out1
		cache = out2
	}
	first1, first2 := out1, out2

	cache, out1, out2 = 0, 0, 0
	position = 0
	counter = ((length - 2) >> 1) + 1
	md51 = word(digest[:], 0) | 1
	md52 = word(digest[:], 4) | 1

	for ; counter > 0; counter-- {
		r0 := word(data, position) + out1
		position += 8
		r1 := r0 * md51
		r1 = 0xB1110000*r1 - 0x30674EEF*(r1>>16)
		r2 := 0x5B9F0000*r1 - 0x78F7A461*(r1>>16)
		r2 = 0x12CEB96D*(r2>>16) - 0x46930000*r2
		r3 := 0x1D830000*r2 + 0x257E1D83*(r2>>16)
		r4 := md52 * (r3 + word(data, position-4))
		r4 = 0x16F50000*r4 - 0x5D8BE90B*(r4>>16)
		r5 := 0x96FF0000*r4 - 0x2C7C6901*(r4>>16)
		r5 = 0x2B890000*r5 + 0x7C932B89*(r5>>16)
		out1 = 0x9F690000*r5 - 0x405B6097*(r5>>16)
		out2 = out1 + cache + r3
		cache = out2
	}

	result := make([]byte, 8)
	binary.LittleEndian.PutUint32(result[0:4], out1^first1)
	binary.LittleEndian.PutUint32(result[4:8], out2^first2)
	return base64.StdEncoding.EncodeToString(result)
}

func word(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

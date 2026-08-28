package promptpay

import "fmt"

const (
	crcPolynomial = 0x1021
	crcInitial    = 0xFFFF
)

func crc16(payload string) string {
	crc := uint16(crcInitial)

	for i := range len(payload) {
		crc ^= uint16(payload[i]) << 8

		for range 8 {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ crcPolynomial
			} else {
				crc <<= 1
			}
		}
	}

	return fmt.Sprintf("%04X", crc)
}

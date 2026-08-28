package promptpay

func CRC16ForTesting(payload string) string {
	return crc16(payload)
}

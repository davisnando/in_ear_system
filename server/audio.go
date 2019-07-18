package main

func audioMix(a int16, b int16) int16 {
	if a+b > 32767 || a+b < -32768 {
		if a > b {
			return a
		} else {
			return b
		}
	}
	return a + b
}

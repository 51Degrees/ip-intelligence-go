package common

import "hash/fnv"

// generateHash generate 32bit hash code for an input string
func GenerateHash(str string) uint32 {
	h := fnv.New32()
	h.Write([]byte(str))
	return h.Sum32()
}

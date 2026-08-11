package auth

import (
	"crypto/sha1"
	"math/big"
)

// ServerHash is the name a login goes by at the session server. Both ends derive
// it from what only they two know, the client to tell Mojang which server it is
// joining and the server to ask Mojang about that same join, so a hash that
// matches is a client that answered this server's encryption request rather than
// somebody else's.
//
// SHA-1 is not a choice here: it is what the client computes, and a different
// digest is a login Mojang has no record of.
//
// The hex is Java's rather than the usual kind. Java reads the digest back as a
// signed two's complement number and prints that, so half of all hashes come out
// negative, with a leading minus and no leading zeroes.
func ServerHash(serverId string, sharedSecret, publicKey []byte) string {
	digest := sha1.New()

	digest.Write([]byte(serverId))
	digest.Write(sharedSecret)
	digest.Write(publicKey)

	sum := digest.Sum(nil)

	value := new(big.Int).SetBytes(sum)

	// A digest with its top bit set is a negative number to Java, which is the
	// same value less one whole turn of the twenty byte odometer.
	if sum[0]&0x80 != 0 {
		value.Sub(value, new(big.Int).Lsh(big.NewInt(1), uint(len(sum))*8))
	}

	return value.Text(16)
}

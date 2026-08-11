package streams

import (
	"crypto/cipher"
)

// cfb8 is a block cipher in 8-bit cipher feedback mode, which is the mode the
// protocol encrypts a connection with.
//
// The standard library has cipher feedback in the full block width only, and
// that is a different cipher: this one advances a byte at a time, encrypting the
// whole register for every single byte of the message and shifting that byte
// into the register afterwards. It costs a block encryption per byte, which is
// what every implementation of this protocol pays.
type cfb8 struct {
	block cipher.Block

	// register holds the last block's worth of ciphertext, whichever end of the
	// cipher produced it, and is what gets encrypted to make the next byte of
	// key stream. It starts as the initialization vector.
	register  []byte
	keyStream []byte

	// decrypt says which side of the cipher this is. Both sides encrypt the
	// register; they differ only in which byte is fed back into it.
	decrypt bool
}

// newCfb8 returns a cipher.Stream over block, keyed to start from iv. The iv is
// copied, since the register it becomes is rewritten for every byte and the
// caller holds the same bytes as the key.
func newCfb8(block cipher.Block, iv []byte, decrypt bool) *cfb8 {
	register := make([]byte, len(iv))
	copy(register, iv)

	return &cfb8{
		block:     block,
		register:  register,
		keyStream: make([]byte, block.BlockSize()),
		decrypt:   decrypt,
	}
}

func (c *cfb8) XORKeyStream(dst, src []byte) {
	for i, in := range src {
		c.block.Encrypt(c.keyStream, c.register)

		out := in ^ c.keyStream[0]

		// The ciphertext byte is what goes back into the register, so which of
		// the two bytes at hand that is depends on the direction.
		feedback := out
		if c.decrypt {
			feedback = in
		}

		copy(c.register, c.register[1:])
		c.register[len(c.register)-1] = feedback

		dst[i] = out
	}
}

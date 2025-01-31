package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
)

type (
	encryptingPrivateKey [ENCRYPTING_KEY_LENGTH]byte
	encryptingPublicKey  [ENCRYPTING_KEY_LENGTH]byte

	EncryptingPrivateKey interface {
		SealMessageFor(recipientPubKey EncryptingPublicKey, msg []byte) ([]byte, error)
		OpenMessageFrom(senderPublicKey EncryptingPublicKey, msgEncrypted []byte) ([]byte, error)
		Bytes() []byte
	}

	EncryptingPublicKey interface {
		Bytes() []byte
	}

	EncryptingKeypair struct {
		EncryptingPrivateKey
		EncryptingPublicKey
	}
)

const (
	ENCRYPTING_KEY_LENGTH   = 32
	ENCRYPTING_NONCE_LENGTH = 24
)

var (
	ErrCannotDecrypt = errors.New("cannot decrypt")
)

func GenerateEncryptingKeypair() (*EncryptingKeypair, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &EncryptingKeypair{
		EncryptingPrivateKey: (*encryptingPrivateKey)(privateKey),
		EncryptingPublicKey:  (*encryptingPublicKey)(publicKey),
	}, nil
}

func EncryptingPublicKeyFromBytes(bs []byte) EncryptingPublicKey {
	var pk encryptingPublicKey
	copy(pk[:], bs)
	return &pk
}

func EncryptingPublicKeyFromHex(s string) (EncryptingPublicKey, error) {
	bs, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var pk encryptingPublicKey
	copy(pk[:], bs)
	return &pk, nil
}

func (pubkey *encryptingPublicKey) Bytes() []byte {
	bs := make([]byte, ENCRYPTING_KEY_LENGTH)
	copy(bs, (*pubkey)[:])
	return bs
}

func EncryptingPrivateKeyFromBytes(bs []byte) EncryptingPrivateKey {
	var pk encryptingPrivateKey
	copy(pk[:], bs)
	return &pk
}

func EncryptingPrivateKeyFromHex(s string) (EncryptingPrivateKey, error) {
	bs, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var pk encryptingPrivateKey
	copy(pk[:], bs)
	return &pk, nil
}

func (privkey *encryptingPrivateKey) Bytes() []byte {
	bs := make([]byte, ENCRYPTING_KEY_LENGTH)
	copy(bs, (*privkey)[:])
	return bs
}

func (privkey *encryptingPrivateKey) SealMessageFor(recipientPubKey EncryptingPublicKey, msg []byte) ([]byte, error) {
	// The shared key can be used to speed up processing when using the same
	// pair of keys repeatedly.
	var sharedEncryptKey [ENCRYPTING_KEY_LENGTH]byte
	box.Precompute(&sharedEncryptKey, (*[ENCRYPTING_KEY_LENGTH]byte)(recipientPubKey.(*encryptingPublicKey)), (*[ENCRYPTING_KEY_LENGTH]byte)(privkey))

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [ENCRYPTING_NONCE_LENGTH]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// This encrypts msg and appends the result to the nonce.
	encrypted := box.SealAfterPrecomputation(nonce[:], msg, &nonce, &sharedEncryptKey)

	return encrypted, nil
}

func (privkey *encryptingPrivateKey) OpenMessageFrom(senderPublicKey EncryptingPublicKey, msgEncrypted []byte) ([]byte, error) {
	// The shared key can be used to speed up processing when using the same
	// pair of keys repeatedly.
	var sharedDecryptKey [ENCRYPTING_KEY_LENGTH]byte
	box.Precompute(&sharedDecryptKey, (*[ENCRYPTING_KEY_LENGTH]byte)(senderPublicKey.(*encryptingPublicKey)), (*[ENCRYPTING_KEY_LENGTH]byte)(privkey))

	// The recipient can decrypt the message using the shared key. When you
	// decrypt, you must use the same nonce you used to encrypt the message.
	// One way to achieve this is to store the nonce alongside the encrypted
	// message. Above, we prefixed the message with the nonce.
	var decryptNonce [ENCRYPTING_NONCE_LENGTH]byte
	copy(decryptNonce[:], msgEncrypted[:ENCRYPTING_NONCE_LENGTH])
	decrypted, ok := box.OpenAfterPrecomputation(nil, msgEncrypted[ENCRYPTING_NONCE_LENGTH:], &decryptNonce, &sharedDecryptKey)
	if !ok {
		return nil, ErrCannotDecrypt
	}
	return decrypted, nil
}

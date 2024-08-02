package httpsign

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
)

// SigningMethodEd25519 implements the EdDSA family.
// Expects ed25519.PrivateKey for signing and ed25519.PublicKey for verification
type SigningMethodEd25519 struct{}

// Specific instance for EdDSA
var SigningMethodEdDSA = &SigningMethodEd25519{}

func (m *SigningMethodEd25519) Alg() string {
	return "EdDSA"
}

// Verify implements token verification for the SigningMethod.
// For this verify method, key must be an ed25519.PublicKey
func (m *SigningMethodEd25519) Verify(signingBytes, sig []byte, key any) error {
	ed25519Key, ok := key.(ed25519.PublicKey)
	if !ok {
		return ErrKeyTypeInvalid
	}
	if len(ed25519Key) != ed25519.PublicKeySize {
		return ErrKeyInvalid
	}

	// Verify the signature
	if !ed25519.Verify(ed25519Key, signingBytes, sig) {
		return ErrSignatureInvalid
	}
	return nil
}

// Sign implements token signing for the SigningMethod.
// For this signing method, key must be an ed25519.PrivateKey
func (m *SigningMethodEd25519) Sign(signingBytes []byte, key any) ([]byte, error) {
	ed25519Key, ok := key.(crypto.Signer)
	if !ok {
		return nil, ErrKeyTypeInvalid
	}

	if _, ok := ed25519Key.Public().(ed25519.PublicKey); !ok {
		return nil, ErrKeyInvalid
	}

	// Sign the string and return the result. ed25519 performs a two-pass hash
	// as part of its algorithm. Therefore, we need to pass a non-prehashed
	// message into the Sign function, as indicated by crypto.Hash(0)
	sig, err := ed25519Key.Sign(rand.Reader, signingBytes, crypto.Hash(0))
	if err != nil {
		return nil, err
	}
	return sig, nil
}

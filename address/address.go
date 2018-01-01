package address

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/frankh/rai"
	"github.com/golang/crypto/blake2b"
	"math"
	"runtime"
	"strings"
	// We've forked golang's ed25519 implementation
	// to use blake2b instead of sha3
	"github.com/frankh/crypto/ed25519"
)

// xrb uses a non-standard base32 character set.
const encodeXrb = "13456789abcdefghijkmnopqrstuwxyz"

var XrbEncoding = base32.NewEncoding(encodeXrb)

func reversed(str []byte) (result []byte) {
	for i := len(str) - 1; i >= 0; i-- {
		result = append(result, str[i])
	}
	return result
}

func ValidateAddress(account rai.Account) bool {
	_, err := AddressToPub(account)

	return err == nil
}

func AddressToPub(account rai.Account) (public_key []byte, err error) {
	address := string(account)
	// A valid xrb address is 64 bytes long
	// First 4 are simply a hard-coded string xrb_ for ease of use
	// The following 52 characters form the address, and the final
	// 8 are a checksum.
	// They are base 32 encoded with a custom encoding.
	if len(address) == 64 && address[:4] == "xrb_" {
		// The xrb address string is 260bits which doesn't fall on a
		// byte boundary. pad with zeros to 280bits.
		// (zeros are encoded as 1 in xrb's 32bit alphabet)
		key_b32xrb := "1111" + address[4:56]
		input_checksum := address[56:]

		key_bytes, err := XrbEncoding.DecodeString(key_b32xrb)
		if err != nil {
			return nil, err
		}
		// strip off upper 24 bits (3 bytes). 20 padding was added by us,
		// 4 is unused as account is 256 bits.
		key_bytes = key_bytes[3:]

		// xrb checksum is calculated by hashing the key and reversing the bytes
		valid := XrbEncoding.EncodeToString(GetAddressChecksum(key_bytes)) == input_checksum
		if valid {
			return key_bytes, nil
		} else {
			return nil, errors.New("Invalid address checksum")
		}
	}

	return nil, errors.New("Invalid address format")
}

func IsValidPrefix(prefix string) bool {
	for _, c := range prefix {
		if !strings.Contains(encodeXrb, string(c)) {
			return false
		}
	}
	return true
}

func EstimatedIterations(prefix string) int {
	return int(math.Pow(32, float64(len(prefix))) / 2)
}

func GenerateVanityAddress(prefix string) (string, rai.Account, error) {
	if !IsValidPrefix(prefix) {
		return "", "", fmt.Errorf("Invalid character in prefix.")
	}

	c := make(chan string, 100)
	progress := make(chan int, 100)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func(c chan string, progress chan int) {
			defer func() {
				recover()
			}()
			count := 0
			for {
				count += 1
				if count%(500+i) == 0 {
					progress <- count
					count = 0
				}
				seed_bytes := make([]byte, 32)
				rand.Read(seed_bytes)
				seed := hex.EncodeToString(seed_bytes)
				pub, _ := KeypairFromSeed(seed, 0)
				address := string(PubKeyToAddress(pub))

				if address[4] != '1' && address[4] != '3' {
					c <- seed
					break
				}

				if strings.HasPrefix(address[5:], prefix) {
					c <- seed
					break
				}
			}
		}(c, progress)
	}

	go func(progress chan int) {
		total := 0
		fmt.Println()
		for {
			count, ok := <-progress
			if !ok {
				break
			}
			total += count
			fmt.Printf("\033[1A\033[KTried %d (~%.2f%%)\n", total, float64(total)/float64(EstimatedIterations(prefix))*100)
		}
	}(progress)

	seed := <-c
	pub, _ := KeypairFromSeed(seed, 0)
	address := PubKeyToAddress(pub)
	if !ValidateAddress(address) {
		return "", "", fmt.Errorf("Address generated had an invalid checksum! Please create an issue on github.")
	}

	close(c)
	close(progress)

	return seed, address, nil
}

func GetAddressChecksum(pub ed25519.PublicKey) []byte {
	hash, err := blake2b.New(5, nil)
	if err != nil {
		panic("Unable to create hash")
	}

	hash.Write(pub)
	return reversed(hash.Sum(nil))
}

func PubKeyToAddress(pub ed25519.PublicKey) rai.Account {
	// Pubkey is 256bits, base32 must be multiple of 5 bits
	// to encode properly.
	// Pad the start with 0's and strip them off after base32 encoding
	padded := append([]byte{0, 0, 0}, pub...)
	address := XrbEncoding.EncodeToString(padded)[4:]
	checksum := XrbEncoding.EncodeToString(GetAddressChecksum(pub))

	return rai.Account("xrb_" + address + checksum)
}

func KeypairFromPrivateKey(private_key string) (ed25519.PublicKey, ed25519.PrivateKey) {
	private_bytes, _ := hex.DecodeString(private_key)
	pub, priv, _ := ed25519.GenerateKey(bytes.NewReader(private_bytes))

	return pub, priv
}

func KeypairFromSeed(seed string, index uint32) (ed25519.PublicKey, ed25519.PrivateKey) {
	// This seems to be the standard way of producing wallets.

	// We hash together the seed with an address index and use
	// that as the private key. Whenever you "add" an address
	// to your wallet the wallet software increases the index
	// and generates a new address.
	hash, err := blake2b.New(32, nil)
	if err != nil {
		panic("Unable to create hash")
	}

	seed_data, err := hex.DecodeString(seed)
	if err != nil {
		panic("Invalid seed")
	}

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, index)

	hash.Write(seed_data)
	hash.Write(bs)

	seed_bytes := hash.Sum(nil)
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seed_bytes))

	if err != nil {
		panic("Unable to generate ed25519 key")
	}

	return pub, priv
}

func GenerateKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	pubkey, privkey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("Unable to generate ed25519 key")
	}

	return pubkey, privkey
}

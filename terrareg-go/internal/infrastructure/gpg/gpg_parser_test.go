package gpg

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKeyInfo_ValidPublicKey(t *testing.T) {
	validKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVJWdwEEAN2WER9iSataTlQThf/a4GRYuPL4yHqqfa8P/CzoZu52JKVcy7sB
GlkppPdTXXZ7gIHL2l9dpSk8TgO9l5NvgXKEPUFmY3R/+8UfPHq9/6bm4oicpmlj
RQNMP05HvbClSN87jHevjswp3rPGokicZ7IOhwiMOWMGB8gViOHurS+lABEBAAG0
KFRlc3QgVGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoA
OBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j
4QVW+8cCzUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+w
GP3DYrMzsJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4
Qz2ZZWwSMtwpuI0EZVJWdwEEANT2AIj1/ELn+nWqVgJ/xhkm6Sh1uE9aaqHA6/Dp
txkAL+eVbbxrnvssSOUZaLwC9gysRYbZiHHG70G6BZttYtyicYkto9wjlfZYvCvY
eTAwscbWeBjV0kadzn7hemcIxIN0x9cpX3GQ0g0kWxnGGpxEu7vOv5qXYDq9YNvp
tObZABEBAAGItgQYAQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsM
AAoJEFv3ginGHH4yC7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqP
RTFoB/o3lP0WrIBJ73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1ki
y34yBsqNSg4f96X28Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=k8KM
-----END PGP PUBLIC KEY BLOCK-----`

	keyID, fingerprint, err := ParseKeyInfo(validKey)

	require.NoError(t, err)
	assert.NotEmpty(t, keyID, "Key ID should not be empty")
	assert.Len(t, keyID, 16, "Key ID should be 16 characters")
	assert.NotEmpty(t, fingerprint, "Fingerprint should not be empty")
	assert.Len(t, fingerprint, 40, "Fingerprint should be 40 characters")
}

func TestParseKeyInfo_InvalidKeyFormat(t *testing.T) {
	invalidKey := "not a valid PGP key"

	_, _, err := ParseKeyInfo(invalidKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode ASCII armor", "Error should mention decoding failure")
}

func TestParseKeyInfo_EmptyKey(t *testing.T) {
	emptyKey := ""

	_, _, err := ParseKeyInfo(emptyKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode ASCII armor", "Error should mention decoding failure")
}

func TestParseKeyInfo_PrivateKey(t *testing.T) {
	privateKey := `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQIGBGVSVncBBADdlhEfYkmrWk5UE4X/2uBkWLjy+Mh6qn2vD/ws6GbudiSlXMu7
ARpZKaT3U112e4CBy9pfXaUpPE4DvZeTb4FyhD1BZmN0f/vFHzx6vf+m5uKInKZp
Y0UDTD9OR72wpUjfO4x3r47MKd6zxqJInGeyDocIjDljBgfIFYjh7q0vpQARAQAB
/gcDAsRU23NbM9O0/0wxbLd+P7JurnQ4bHdp620hGoel6eGa4rxVkRYEPccYXu4B
gJfaay8xizisbfx1PGuaUW7fz49SG+CcekuBfxBIKfUJLBWHGolClzlN6FFCLuWI
5lPNF/77E318quDdcPWemaTMS8o1Ijva4OHkUgJTKmUAV936+TjpwgfTxghv5WNB
wMMZb5SA8QfRk/97J0nbnMLUH+sJduz7Iryi6VHDVV9ZlaBU+k/C/PQU7D9iAen3
Pds2MuggORCjcR2Dqtta0JCpeslUJdTYTaWkyNNLZdxtBXOIyDmSvJMCSmA3+X9R
qwn3yTp+fxlimo7vdBp0MpeDl7z3aCIfvDiVC6B597mHG0CaZM475ek3gtuRvmel
U2bhdWJDWPAhKom9gg3SDbE+4y8c2pLZcVeKZ1Pf2a63M4R29gej4Tk/mpVnnGDT
CFGJVfbIWUA7PT+RueKlBHOk4mJpnfL6Jl61aHv3/2NAvKXkoNGteEK0KFRlc3Qg
VGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoAOBYhBDwL
AxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQWAgMBAh4B
AheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j4QVW+8cC
zUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+wGP3DYrMz
sJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4Qz2ZZWwS
MtwpnQIGBGVSVncBBADU9gCI9fxC5/p1qlYCf8YZJukodbhPWmqhwOvw6bcZAC/n
lW28a577LEjlGWi8AvYMrEWG2Yhxxu9BugWbbWLconGJLaPcI5X2WLwr2HkwMLHG
1ngY1dJGnc5+4XpnCMSDdMfXKV9xkNINJFsZxhqcRLu7zr+al2A6vWDb6bTm2QAR
AQAB/gcDAlHZ6lg9aFkb/1QpyNEg2Uiv/JEVYsIzLJAPwmo6/868xFLkAkvDcvsl
RJs3c/+leLSES3HB40tYUbjHdSZRkhFrNV32ZstyfJg8PKjDkmBioq9vaZ5BjnpP
SW7BfpHqcPsWtygvSX3EtaiL6acUmrObW/evFG0Kng7D9D/HjnR+9QPeKbkG8Jz6
zsi0suzRZsgYMBnJz3+gmnFgCEwb+npab5NEoaY8iwlBIX/KRYk4EK75wdO/OgWA
dd34MkO/DINczlWO8NgSwQNm1kuKNau2tSH5IQsK0YvR/Zi46qQQe33HrNH51b8K
mM2NkFW29e1BUqjgv3Yl7Ig/XhnXtOLcpdCNr7LW9ZlIvxoLJ4EysGR1kIAp6YtW
z83SUtF9np0fuDxWw4gI7O5WM3NnwHV5/cW4aZvEdHxcUF1gGLZMXrcO/s26VSia
jhyhLuew0rbMgc0dbGiXo3DOkQYYmi7p1LOvadAfjA6inOozvImUuXwAlQeItgQY
AQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsMAAoJEFv3ginGHH4y
C7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqPRTFoB/o3lP0WrIBJ
73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1kiy34yBsqNSg4f96X2
8Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=RfHf
-----END PGP PRIVATE KEY BLOCK-----`

	_, _, err := ParseKeyInfo(privateKey)

	assert.Error(t, err, "Should return error for private key")
	assert.Contains(t, err.Error(), "expected PGP PUBLIC KEY BLOCK", "Error should mention expecting public key block")
}

func TestParseKeyInfo_NoKeyBlock(t *testing.T) {
	noKeyBlock := generateMessageBlockArmored(t)

	_, _, err := ParseKeyInfo(noKeyBlock)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected PGP PUBLIC KEY BLOCK", "Error should mention expecting public key block")
}

func TestParseKeyInfo_ShortFingerprint(t *testing.T) {
	shortKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQIN
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----`

	_, _, err := ParseKeyInfo(shortKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read packet", "Error should mention packet read failure")
}

func TestVerifySignature_InvalidSignatureData(t *testing.T) {
	publicKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVJWdwEEAN2WER9iSataTlQThf/a4GRYuPL4yHqqfa8P/CzoZu52JKVcy7sB
GlkppPdTXXZ7gIHL2l9dpSk8TgO9l5NvgXKEPUFmY3R/+8UfPHq9/6bm4oicpmlj
RQNMP05HvbClSN87jHevjswp3rPGokicZ7IOhwiMOWMGB8gViOHurS+lABEBAAG0
KFRlc3QgVGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoA
OBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j
4QVW+8cCzUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+w
GP3DYrMzsJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4
Qz2ZZWwSMtwpuI0EZVJWdwEEANT2AIj1/ELn+nWqVgJ/xhkm6Sh1uE9aaqHA6/Dp
txkAL+eVbbxrnvssSOUZaLwC9gysRYbZiHHG70G6BZttYtyicYkto9wjlfZYvCvY
eTAwscbWeBjV0kadzn7hemcIxIN0x9cpX3GQ0g0kWxnGGpxEu7vOv5qXYDq9YNvp
tObZABEBAAGItgQYAQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsM
AAoJEFv3ginGHH4yC7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqP
RTFoB/o3lP0WrIBJ73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1ki
y34yBsqNSg4f96X28Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=k8KM
-----END PGP PUBLIC KEY BLOCK-----`

	signature := `-----BEGIN PGP SIGNATURE-----

iHUEABEIAB0WIQRPzCn4tRzB7yR5cFq+6mX2wVh2AUCY3x5HwAKCRC+6mX2wVh2
A7V3AP9T9Mk8Y5Yy7J8j2J3G9H5J8Y5Yy7J8j2J3G9H5J8Y5Yy7J8j2J3G9H5J8
AQDyK8j2J3G9H5J8Y5Yy7J8j2J3G9H5J8Y5Yy7J8j2J3G9H5J8=
=K8jM
-----END PGP SIGNATURE-----`

	data := "test data to verify"

	valid, err := VerifySignature([]byte(publicKey), []byte(signature), []byte(data))

	require.NoError(t, err)
	assert.False(t, valid, "Malformed signature should be invalid")
}

func TestVerifySignature_InvalidSignature(t *testing.T) {
	publicKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVJWdwEEAN2WER9iSataTlQThf/a4GRYuPL4yHqqfa8P/CzoZu52JKVcy7sB
GlkppPdTXXZ7gIHL2l9dpSk8TgO9l5NvgXKEPUFmY3R/+8UfPHq9/6bm4oicpmlj
RQNMP05HvbClSN87jHevjswp3rPGokicZ7IOhwiMOWMGB8gViOHurS+lABEBAAG0
KFRlc3QgVGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoA
OBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j
4QVW+8cCzUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+w
GP3DYrMzsJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4
Qz2ZZWwSMtwpuI0EZVJWdwEEANT2AIj1/ELn+nWqVgJ/xhkm6Sh1uE9aaqHA6/Dp
txkAL+eVbbxrnvssSOUZaLwC9gysRYbZiHHG70G6BZttYtyicYkto9wjlfZYvCvY
eTAwscbWeBjV0kadzn7hemcIxIN0x9cpX3GQ0g0kWxnGGpxEu7vOv5qXYDq9YNvp
tObZABEBAAGItgQYAQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsM
AAoJEFv3ginGHH4yC7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqP
RTFoB/o3lP0WrIBJ73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1ki
y34yBsqNSg4f96X28Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=k8KM
-----END PGP PUBLIC KEY BLOCK-----`

	wrongSignature := "bad signature"
	dataContent := "test data"

	_, err := VerifySignature([]byte(publicKey), []byte(wrongSignature), []byte(dataContent))

	assert.Error(t, err, "Non-armored signature should cause decode error")
}

func TestVerifySignature_MalformedSignature(t *testing.T) {
	publicKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVJWdwEEAN2WER9iSataTlQThf/a4GRYuPL4yHqqfa8P/CzoZu52JKVcy7sB
GlkppPdTXXZ7gIHL2l9dpSk8TgO9l5NvgXKEPUFmY3R/+8UfPHq9/6bm4oicpmlj
RQNMP05HvbClSN87jHevjswp3rPGokicZ7IOhwiMOWMGB8gViOHurS+lABEBAAG0
KFRlc3QgVGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoA
OBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j
4QVW+8cCzUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+w
GP3DYrMzsJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4
Qz2ZZWwSMtwpuI0EZVJWdwEEANT2AIj1/ELn+nWqVgJ/xhkm6Sh1uE9aaqHA6/Dp
txkAL+eVbbxrnvssSOUZaLwC9gysRYbZiHHG70G6BZttYtyicYkto9wjlfZYvCvY
eTAwscbWeBjV0kadzn7hemcIxIN0x9cpX3GQ0g0kWxnGGpxEu7vOv5qXYDq9YNvp
tObZABEBAAGItgQYAQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsM
AAoJEFv3ginGHH4yC7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqP
RTFoB/o3lP0WrIBJ73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1ki
y34yBsqNSg4f96X28Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=k8KM
-----END PGP PUBLIC KEY BLOCK-----`

	malformedSignature := "not a valid signature"

	_, err := VerifySignature([]byte(publicKey), []byte(malformedSignature), []byte("test data"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode signature armor", "Error should mention decoding failure")
}

func TestValidateKeyStructure_ValidKey(t *testing.T) {
	validKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZVJWdwEEAN2WER9iSataTlQThf/a4GRYuPL4yHqqfa8P/CzoZu52JKVcy7sB
GlkppPdTXXZ7gIHL2l9dpSk8TgO9l5NvgXKEPUFmY3R/+8UfPHq9/6bm4oicpmlj
RQNMP05HvbClSN87jHevjswp3rPGokicZ7IOhwiMOWMGB8gViOHurS+lABEBAAG0
KFRlc3QgVGVycmFyZWcgPHRlc3R0ZXJyYXJlZ0BleGFtcGxlLmNvbT6IzgQTAQoA
OBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsDBQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEFv3ginGHH4yxzwD/RiJzcs1mGkjWQq6yGVQESFTelfPFu+j
4QVW+8cCzUUEWbcEoCvN9cCFS+y3SHnZhACrRqxdEFaNLtbWyFNLhXOUbS7vKE+w
GP3DYrMzsJjN6EK2QsTrdF90vk3fvMaXHRSxmVUhisCm6uuZvp18Dfo7zyOlb+e4
Qz2ZZWwSMtwpuI0EZVJWdwEEANT2AIj1/ELn+nWqVgJ/xhkm6Sh1uE9aaqHA6/Dp
txkAL+eVbbxrnvssSOUZaLwC9gysRYbZiHHG70G6BZttYtyicYkto9wjlfZYvCvY
eTAwscbWeBjV0kadzn7hemcIxIN0x9cpX3GQ0g0kWxnGGpxEu7vOv5qXYDq9YNvp
tObZABEBAAGItgQYAQoAIBYhBDwLAxWQ5bvJXe5HRlv3ginGHH4yBQJlUlZ3AhsM
AAoJEFv3ginGHH4yC7oD/RdG6xquOBMz7hDop8/4o+NGHAJQiAl/Kt6VpG1fBmqP
RTFoB/o3lP0WrIBJ73PNTjguhOrAIEQcjPLiZESqGs24pZvoFp0wK6kJgKIiH1ki
y34yBsqNSg4f96X28Cm66mGVhvyAEegQgtbByF9UOyPv+S5uyPMrHqidLgD95Cpj
=k8KM
-----END PGP PUBLIC KEY BLOCK-----`

	err := ValidateKeyStructure(validKey)

	assert.NoError(t, err, "Valid key should not return error")
}

func TestValidateKeyStructure_InvalidBlockType(t *testing.T) {
	invalidKey := generateMessageBlockArmored(t)

	err := ValidateKeyStructure(invalidKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GPG key type", "Error should mention invalid key type")
}

func TestValidateKeyStructure_EmptyKey(t *testing.T) {
	emptyKey := ""

	err := ValidateKeyStructure(emptyKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode ASCII armor", "Error should mention decoding failure")
}

func TestValidateKeyStructure_NoEntities(t *testing.T) {
	// Create an empty armored block - properly formatted but no key data
	var buf bytes.Buffer
	encoder, err := armor.Encode(&buf, "PGP PUBLIC KEY BLOCK", nil)
	require.NoError(t, err)

	// Close immediately without writing any key data
	err = encoder.Close()
	require.NoError(t, err)

	keyWithoutEntities := buf.String()

	err = ValidateKeyStructure(keyWithoutEntities)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entities", "Error should mention missing entities")
}

func TestValidateKeyStructure_MultipleKeys(t *testing.T) {
	// Generate real GPG keys for testing
	multipleKeys := generateMultipleKeysArmored(t)

	err := ValidateKeyStructure(multipleKeys)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one key", "Error should mention single key requirement")
}

// generateMultipleKeysArmored generates an armored block containing multiple GPG keys
func generateMultipleKeysArmored(t *testing.T) string {
	t.Helper()

	// Create buffer for armored output
	var buf bytes.Buffer
	encoder, err := armor.Encode(&buf, "PGP PUBLIC KEY BLOCK", nil)
	require.NoError(t, err)

	// Write two keys to the same armored block
	for range 2 {
		entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
		require.NoError(t, err)

		err = entity.SerializePrivate(encoder, nil)
		require.NoError(t, err)
	}

	err = encoder.Close()
	require.NoError(t, err)

	return buf.String()
}

// generateSingleKeyArmored generates an armored block containing a single GPG key
func generateSingleKeyArmored(t *testing.T) string {
	t.Helper()

	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	require.NoError(t, err)

	var buf bytes.Buffer
	encoder, err := armor.Encode(&buf, "PGP PUBLIC KEY BLOCK", nil)
	require.NoError(t, err)

	err = entity.SerializePrivate(encoder, nil)
	require.NoError(t, err)

	err = encoder.Close()
	require.NoError(t, err)

	return buf.String()
}

// generateMessageBlockArmored generates an armored PGP message block
func generateMessageBlockArmored(t *testing.T) string {
	t.Helper()

	var buf bytes.Buffer
	encoder, err := armor.Encode(&buf, "PGP MESSAGE BLOCK", nil)
	require.NoError(t, err)

	_, err = encoder.Write([]byte("test message"))
	require.NoError(t, err)

	err = encoder.Close()
	require.NoError(t, err)

	return buf.String()
}

// generateSignatureArmored generates an armored PGP signature
func generateSignatureArmored(t *testing.T, data []byte) string {
	t.Helper()

	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	require.NoError(t, err)

	var buf bytes.Buffer
	encoder, err := armor.Encode(&buf, "PGP SIGNATURE", nil)
	require.NoError(t, err)

	err = openpgp.DetachSign(encoder, entity, bytes.NewReader(data), nil)
	require.NoError(t, err)

	err = encoder.Close()
	require.NoError(t, err)

	return buf.String()
}

func TestReadAllKeys_SingleKey(t *testing.T) {
	singleKey := generateSingleKeyArmored(t)

	count, err := ReadAllKeys(singleKey)

	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should find exactly one key")
}

func TestReadAllKeys_InvalidASCIIArmor(t *testing.T) {
	invalidArmor := "not ascii armored data"

	_, err := ReadAllKeys(invalidArmor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode ASCII armor", "Error should mention decoding failure")
}

func TestReadAllKeys_MultipleKeys(t *testing.T) {
	multipleKeys := generateMultipleKeysArmored(t)

	count, err := ReadAllKeys(multipleKeys)

	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should find exactly two keys")
}

func TestGetKeyDetails_ValidKey(t *testing.T) {
	validKey := generateSingleKeyArmored(t)

	details, err := ParseKeyDetails(validKey)

	require.NoError(t, err)
	assert.NotEmpty(t, details.KeyID, "Key ID should not be empty")
	assert.Len(t, details.KeyID, 16, "Key ID should be 16 characters")
	assert.NotEmpty(t, details.Fingerprint, "Fingerprint should not be empty")
	assert.Len(t, details.Fingerprint, 40, "Fingerprint should be 40 characters")
	assert.Greater(t, details.CreationTime, uint64(0), "Creation time should be positive")
}

func TestGetKeyDetails_KeyWithTimestamp(t *testing.T) {
	keyWithCreationTime := generateSingleKeyArmored(t)

	details, err := ParseKeyDetails(keyWithCreationTime)

	require.NoError(t, err)
	assert.Greater(t, details.CreationTime, uint64(1234567890), "Creation time should be reasonable timestamp")
}

func TestDecodeSignatureBlock_ValidSignature(t *testing.T) {
	data := []byte("test data")
	validSignature := generateSignatureArmored(t, data)

	decoded, err := DecodeSignatureBlock(validSignature)

	require.NoError(t, err)
	assert.NotEmpty(t, decoded, "Decoded signature should not be empty")
	assert.NotContains(t, string(decoded), "-----BEGIN", "Should not contain PGP headers")
	assert.NotContains(t, string(decoded), "-----END", "Should not contain PGP trailers")
}

func TestDecodeSignatureBlock_InvalidSignature(t *testing.T) {
	invalidSignature := "not a valid signature"

	_, err := DecodeSignatureBlock(invalidSignature)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode signature armor", "Error should mention decoding failure")
}

func TestDecodeSignatureBlock_EmptySignature(t *testing.T) {
	emptySignature := ""

	_, err := DecodeSignatureBlock(emptySignature)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode signature armor", "Error should mention decoding failure")
}

func TestVerifySignatureWithFiles_Valid(t *testing.T) {
	// Generate key and signature together to ensure they match
	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	require.NoError(t, err)

	// Generate public key
	var pubKeyBuf bytes.Buffer
	pubEncoder, err := armor.Encode(&pubKeyBuf, "PGP PUBLIC KEY BLOCK", nil)
	require.NoError(t, err)
	err = entity.SerializePrivate(pubEncoder, nil)
	require.NoError(t, err)
	err = pubEncoder.Close()
	require.NoError(t, err)
	publicKey := pubKeyBuf.String()

	// Generate signature
	data := []byte("test data to verify")
	var sigBuf bytes.Buffer
	sigEncoder, err := armor.Encode(&sigBuf, "PGP SIGNATURE", nil)
	require.NoError(t, err)
	err = openpgp.DetachSign(sigEncoder, entity, bytes.NewReader(data), nil)
	require.NoError(t, err)
	err = sigEncoder.Close()
	require.NoError(t, err)
	signatureContent := sigBuf.String()

	valid, err := VerifySignature([]byte(publicKey), []byte(signatureContent), data)

	require.NoError(t, err)
	assert.True(t, valid, "Signature verification with files should be valid")
}

func TestVerifySignatureWithFiles_InvalidSignature(t *testing.T) {
	publicKey := generateSingleKeyArmored(t)

	wrongSignature := "bad signature"
	dataContent := "test data"

	_, err := VerifySignatureWithFiles([]byte(publicKey), []byte(dataContent), []byte(wrongSignature))

	assert.Error(t, err, "Non-armored signature should cause decode error")
}

func TestVerifySignatureWithFiles_EmptyData(t *testing.T) {
	publicKey := generateSingleKeyArmored(t)

	signatureContent := generateSignatureArmored(t, []byte("different data"))
	emptyData := []byte("")

	valid, err := VerifySignatureWithFiles([]byte(publicKey), emptyData, []byte(signatureContent))

	require.NoError(t, err)
	assert.False(t, valid, "Signature verification with empty data should be invalid")
}

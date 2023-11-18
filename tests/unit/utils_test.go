package unit_tests

import (
	"bytes"
	"encoding/base64"
	"math/rand"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/internal/utils/compression"
	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/argon2"
)

// GetStringHashWithArgon uses Argon2 version 0x13 to hash a plaintext password with a provided salt string and return hash as base64 string.
func getStringHashWithArgon(passphrase, salt string, timeConsideration uint32, threads uint8, hashLength uint32) string {

	if passphrase == "" || salt == "" {
		return ""
	}

	if timeConsideration == 0 {
		timeConsideration = 1
	}

	if threads == 0 {
		threads = 1
	}

	hashy := argon2.IDKey([]byte(passphrase), []byte(salt), timeConsideration, 64*1024, threads, hashLength)

	base64Hash := make([]byte, base64.StdEncoding.EncodedLen(len(hashy)))
	base64.StdEncoding.Encode(base64Hash, hashy)

	return string(base64Hash)
}

func TestCompressAndDecompressWithGzip(t *testing.T) {
	data := "SuperStreetFighter2TurboMBisonDidNothingWrong"
	buffer := &bytes.Buffer{}

	err := compression.Compress("gzip", []byte(data), buffer)
	assert.NoError(t, err)

	assert.NotEqual(t, nil, buffer)
	assert.NotEqual(t, 0, buffer.Len())

	err = compression.Decompress("gzip", buffer)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, buffer)
	assert.Equal(t, data, buffer.String())
}

func TestCompressAndDecompressWithZstd(t *testing.T) {
	data := "SuperStreetFighter2TurboMBisonDidNothingWrong"
	buffer := &bytes.Buffer{}

	err := compression.Compress("zstd", []byte(data), buffer)
	assert.NoError(t, err)

	assert.NotEqual(t, nil, buffer)
	assert.NotEqual(t, 0, buffer.Len())

	err = compression.Decompress("zstd", buffer)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, buffer)
	assert.Equal(t, data, buffer.String())
}

func TestGetHashWithArgon2(t *testing.T) {

	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 64)
	assert.NotNil(t, hashy)

	t.Logf("Password: %s\tLength: %d\r\n", password, len(password))
	t.Logf("Hashed Password: %s\tLength: %d\r\n", hashy, len(hashy))

	base64Hash := make([]byte, base64.StdEncoding.EncodedLen(len(hashy)))
	base64.StdEncoding.Encode(base64Hash, hashy)

	t.Logf("Hashed As Base64: %s\tLength: %d\r\n", base64Hash, len(base64Hash))
}

func TestGetHashStringWithArgon2(t *testing.T) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	base64Hash := getStringHashWithArgon(password, salt, 1, 12, 64)
	assert.NotNil(t, base64Hash)

	t.Logf("Password: %s\tLength: %d\r\n", password, len(password))
	t.Logf("Hash As Base64: %s\tLength: %d\r\n", base64Hash, len(base64Hash))
}

func TestHashAndAesEncrypt(t *testing.T) {
	dataPayload := []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)
	assert.NotNil(t, hashy)

	base64Hash := make([]byte, base64.StdEncoding.EncodedLen(len(hashy)))
	base64.StdEncoding.Encode(base64Hash, hashy)

	t.Logf("Password: %s\tLength: %d\r\n", password, len(password))
	t.Logf("Password Hashed As Base64: %s\tLength: %d\r\n", base64Hash, len(base64Hash))

	buffer := &bytes.Buffer{}
	err := crypto.Encrypt("aes", hashy, dataPayload, buffer)
	assert.NoError(t, err)

	base64Encrypted := make([]byte, base64.StdEncoding.EncodedLen(len(buffer.Bytes())))
	base64.StdEncoding.Encode(base64Encrypted, buffer.Bytes())

	t.Logf("Secret Payload: %s\r\n", string(dataPayload))
	t.Logf("Encrypted As Base64: %s\tLength: %d\r\n", base64Encrypted, len(base64Encrypted))
}

func TestHashAndAesEncryptAndDecrypt(t *testing.T) {
	dataPayload := []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)
	assert.NotNil(t, hashy)

	base64Hash := make([]byte, base64.StdEncoding.EncodedLen(len(hashy)))
	base64.StdEncoding.Encode(base64Hash, hashy)

	t.Logf("Password: %s\tLength: %d\r\n", password, len(password))
	t.Logf("Password Hashed As Base64: %s\tLength: %d\r\n", base64Hash, len(base64Hash))

	buffer := &bytes.Buffer{}
	err := crypto.Encrypt("aes", hashy, dataPayload, buffer)
	assert.NoError(t, err)

	base64Encrypted := make([]byte, base64.StdEncoding.EncodedLen(len(buffer.Bytes())))
	base64.StdEncoding.Encode(base64Encrypted, buffer.Bytes())

	t.Logf("Original Payload: %s\r\n", string(dataPayload))
	t.Logf("Encrypted As Base64: %s\tLength: %d\r\n", base64Encrypted, len(base64Encrypted))

	err = crypto.Decrypt("aes", hashy, buffer)
	assert.NoError(t, err)
	t.Logf("Decrypted Payload: %s\r\n", buffer.String())

	assert.Equal(t, dataPayload, buffer.Bytes())
}

type TestStruct struct {
	PropertyString1 string `json:"PropertyString1"`
	PropertyString2 string `json:"PropertyString2"`
	PropertyString3 string `json:"PropertyString3"`
	PropertyString4 string `json:"PropertyString4"`
}

func TestCreateAndReadCompressedPayload(t *testing.T) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)

	encrypt := &fr.EncryptionConfig{
		Enabled:           false,
		Hashkey:           hashy,
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}

	compression := &fr.CompressionConfig{
		Enabled: true,
		Type:    compression.GzipType,
	}

	test := &TestStruct{
		PropertyString1: mock.RandomString(5000),
		PropertyString2: mock.RandomString(5000),
		PropertyString3: mock.RandomString(5000),
		PropertyString4: mock.RandomString(5000),
	}

	t.Logf("Test Data Size: ~ %d letters", 20000)
	data, err := fr.CreatePayload(test, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
	t.Logf("Compressed Payload Size: ~ %d bytes", len(data))

	buffer := bytes.NewBuffer(data)
	err = fr.ReadPayload(buffer, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buffer.Len())

	var json = jsoniter.ConfigFastest
	outputData := &TestStruct{}
	err = json.Unmarshal(buffer.Bytes(), outputData)
	assert.NoError(t, err)
	assert.Equal(t, test.PropertyString1, outputData.PropertyString1)
	assert.Equal(t, test.PropertyString2, outputData.PropertyString2)
	assert.Equal(t, test.PropertyString3, outputData.PropertyString3)
	assert.Equal(t, test.PropertyString4, outputData.PropertyString4)
}

func TestCreateAndReadEncryptedPayload(t *testing.T) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)

	encrypt := &fr.EncryptionConfig{
		Enabled:           true,
		Hashkey:           hashy,
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}

	compression := &fr.CompressionConfig{
		Enabled: false,
		Type:    compression.GzipType,
	}

	test := &TestStruct{
		PropertyString1: mock.RandomString(5000),
		PropertyString2: mock.RandomString(5000),
		PropertyString3: mock.RandomString(5000),
		PropertyString4: mock.RandomString(5000),
	}

	t.Logf("Test Data Size: ~ %d letters", 20000)
	data, err := fr.CreatePayload(test, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
	t.Logf("Encrypted Payload Size: ~ %d bytes", len(data))

	buffer := bytes.NewBuffer(data)
	err = fr.ReadPayload(buffer, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buffer.Len())

	var json = jsoniter.ConfigFastest
	outputData := &TestStruct{}
	err = json.Unmarshal(buffer.Bytes(), outputData)
	assert.NoError(t, err)
	assert.Equal(t, test.PropertyString1, outputData.PropertyString1)
	assert.Equal(t, test.PropertyString2, outputData.PropertyString2)
	assert.Equal(t, test.PropertyString3, outputData.PropertyString3)
	assert.Equal(t, test.PropertyString4, outputData.PropertyString4)
}

func TestCreateAndReadCompressedEncryptedPayload(t *testing.T) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)

	encrypt := &fr.EncryptionConfig{
		Enabled:           true,
		Hashkey:           hashy,
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}

	compression := &fr.CompressionConfig{
		Enabled: true,
		Type:    compression.GzipType,
	}

	test := &TestStruct{
		PropertyString1: mock.RandomString(5000),
		PropertyString2: mock.RandomString(5000),
		PropertyString3: mock.RandomString(5000),
		PropertyString4: mock.RandomString(5000),
	}

	t.Logf("Test Data Size: ~ %d letters", 20000)
	data, err := fr.CreatePayload(test, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
	t.Logf("Compressed & Encrypted Payload Size: ~ %d bytes", len(data))

	buffer := bytes.NewBuffer(data)
	err = fr.ReadPayload(buffer, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buffer.Len())

	var json = jsoniter.ConfigFastest
	outputData := &TestStruct{}
	err = json.Unmarshal(buffer.Bytes(), outputData)
	assert.NoError(t, err)
	assert.Equal(t, test.PropertyString1, outputData.PropertyString1)
	assert.Equal(t, test.PropertyString2, outputData.PropertyString2)
	assert.Equal(t, test.PropertyString3, outputData.PropertyString3)
	assert.Equal(t, test.PropertyString4, outputData.PropertyString4)
}

func TestCreateAndReadLZCompressedEncryptedPayload(t *testing.T) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	hashy := crypto.GetHashWithArgon(password, salt, 1, 12, 64, 32)

	encrypt := &fr.EncryptionConfig{
		Enabled:           true,
		Hashkey:           hashy,
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}

	compression := &fr.CompressionConfig{
		Enabled: true,
		Type:    compression.ZstdType,
	}

	test := &TestStruct{
		PropertyString1: mock.RandomString(5000),
		PropertyString2: mock.RandomString(5000),
		PropertyString3: mock.RandomString(5000),
		PropertyString4: mock.RandomString(5000),
	}

	t.Logf("Test Data Size: ~ %d letters", 20000)
	data, err := fr.CreatePayload(test, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
	t.Logf("Compressed & Encrypted Payload Size: ~ %d bytes", len(data))

	buffer := bytes.NewBuffer(data)
	err = fr.ReadPayload(buffer, compression, encrypt)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, buffer.Len())

	var json = jsoniter.ConfigFastest
	outputData := &TestStruct{}
	err = json.Unmarshal(buffer.Bytes(), outputData)
	assert.NoError(t, err)
	assert.Equal(t, test.PropertyString1, outputData.PropertyString1)
	assert.Equal(t, test.PropertyString2, outputData.PropertyString2)
	assert.Equal(t, test.PropertyString3, outputData.PropertyString3)
	assert.Equal(t, test.PropertyString4, outputData.PropertyString4)
}

func TestRandomString(t *testing.T) {
	randoString := mock.RandomString(20)
	assert.NotEqual(t, "", randoString)
	t.Logf("RandoString1: %s", randoString)

	time.Sleep(1 * time.Nanosecond)

	anotherRandoString := mock.RandomString(20)
	assert.NotEqual(t, "", anotherRandoString)
	t.Logf("RandoString2: %s", anotherRandoString)

	assert.NotEqual(t, randoString, anotherRandoString)
}

func TestRandomStringFromSource(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())

	randoString := mock.RandomStringFromSource(10, src)
	assert.NotEqual(t, "", randoString)
	t.Logf("RandoString1: %s", randoString)

	anotherRandoString := mock.RandomStringFromSource(10, src)
	assert.NotEqual(t, "", anotherRandoString)
	t.Logf("RandoString2: %s", anotherRandoString)

	assert.NotEqual(t, randoString, anotherRandoString)
}

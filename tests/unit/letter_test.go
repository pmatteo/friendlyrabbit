package unit_tests

import (
	"testing"

	"github.com/google/uuid"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/internal/utils/compression"
	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestLetterCreateDefault(t *testing.T) {
	l, err := fr.NewLetter("Exchange", "Key", []byte("TestUnitQueue"))
	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.NotNil(t, l.LetterID)
	assert.NotEqual(t, uuid.Nil, l.LetterID)
	assert.Equal(t, []byte("TestUnitQueue"), l.Body)

	opts := l.Options()

	assert.NotNil(t, opts)
	assert.Empty(t, opts.Headers)
	assert.Equal(t, "Exchange", opts.Exchange)
	assert.Equal(t, "Key", opts.RoutingKey)
	assert.Equal(t, false, opts.Mandatory)
	assert.Equal(t, false, opts.Immediate)
	assert.Equal(t, "application/json", opts.ContentType)
	assert.Equal(t, "", opts.CorrelationID)
	assert.Equal(t, "", opts.MessageType)
	assert.Equal(t, uint8(0), opts.Priority)
	assert.Equal(t, uint8(2), opts.DeliveryMode)
	assert.Equal(t, uint16(3), opts.RetryCount)
	assert.Nil(t, opts.EConf)
	assert.Nil(t, opts.CConf)
}

func TestLetterCreateWithOptions(t *testing.T) {
	e := &fr.EncryptionConfig{
		Enabled:           false,
		Hashkey:           crypto.GetHashWithArgon("password", "salt", 1, 12, 64, 32),
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}
	c := &fr.CompressionConfig{Enabled: true}
	h := map[string]interface{}{"test-header": "test-value"}

	l, err := fr.NewLetter(
		"Exchange", "Key", []byte("TestUnitQueue"),
		fr.WithHeaders(h),
		fr.WithCorrelationID("correlation-id"),
		fr.WithCompressionConfig(c),
		fr.WithEncryptionConfig(e),
		func(lo *fr.LetterOpts) {
			lo.ContentType = "application/plain-text"
			lo.Priority = 2
			lo.DeliveryMode = 1
			lo.RetryCount = 10
		},
	)

	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.NotNil(t, l.LetterID)
	assert.NotEqual(t, uuid.Nil, l.LetterID)
	assert.NotEmpty(t, l.Body)

	opts := l.Options()

	assert.NotNil(t, opts)
	assert.NotEmpty(t, opts.Headers)
	assert.Equal(t, 1, len(opts.Headers))
	v, ok := opts.Headers["test-header"]
	assert.True(t, ok)
	assert.Equal(t, "test-value", v)

	assert.Equal(t, "Exchange", opts.Exchange)
	assert.Equal(t, "Key", opts.RoutingKey)
	assert.Equal(t, false, opts.Mandatory)
	assert.Equal(t, false, opts.Immediate)
	assert.Equal(t, "application/plain-text", opts.ContentType)
	assert.Equal(t, "correlation-id", opts.CorrelationID)
	assert.Equal(t, "", opts.MessageType)
	assert.Equal(t, uint8(2), opts.Priority)
	assert.Equal(t, uint8(1), opts.DeliveryMode)
	assert.Equal(t, uint16(10), opts.RetryCount)
	assert.NotNil(t, opts.EConf)
	assert.Equal(t, e, opts.EConf)
	assert.NotNil(t, opts.CConf)
	assert.Equal(t, c, opts.CConf)
}

func TestTableNewLetterWithPayload(t *testing.T) {
	encrypt := &fr.EncryptionConfig{
		Enabled:           false,
		Hashkey:           crypto.GetHashWithArgon("password", "salt", 1, 12, 64, 32),
		Type:              crypto.AesSymmetricType,
		TimeConsideration: 1,
		Threads:           6,
	}
	compress := &fr.CompressionConfig{Enabled: true, Type: compression.GzipType}

	data := mock.BasicStruct{N: 1, S: "TestUnitQueue"}

	expected, err := fr.ToPayload(data, nil, nil)
	assert.NoError(t, err)
	expectedCompressed, err := fr.ToPayload(data, compress, nil)
	assert.NoError(t, err)
	expectedEncrypted, err := fr.ToPayload(data, nil, encrypt)
	assert.NoError(t, err)
	expectedComprEncrypt, err := fr.ToPayload(data, compress, encrypt)
	assert.NoError(t, err)

	var tests = []struct {
		name         string
		data         mock.BasicStruct
		expectedData []byte
		encryption   *fr.EncryptionConfig
		compression  *fr.CompressionConfig
	}{
		{"Basic", data, expected, nil, nil},
		{"With compression", data, expectedCompressed, nil, compress},
		{"With encryption", data, expectedEncrypted, encrypt, nil},
		{"With encryption and compression", data, expectedComprEncrypt, encrypt, compress},
	}

	// The execution loop
	for _, tt := range tests {
		data := tt.data
		expectedData := tt.expectedData
		encryption := tt.encryption
		compression := tt.compression

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l, err := fr.NewLetter("Exchange", "Key", data, fr.WithCompressionConfig(compression), fr.WithEncryptionConfig(encryption))

			assert.NotNil(t, l)
			assert.NoError(t, err)

			assert.NotNil(t, l.LetterID)
			assert.NotEmpty(t, l.LetterID)
			assert.NotEqual(t, uuid.Nil, l.LetterID)

			assert.NoError(t, err)
			assert.Equal(t, expectedData, l.Body)
		})
	}
}

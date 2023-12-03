package unit_tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/internal/utils/compression"
	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestEnvelopeCreate(t *testing.T) {
	e := fr.NewEnvelope(context.Background(), "TestUnitExchange", "TestUnitKey", amqp.Table{
		"TestUnitHeader": "TestUnitValue",
	})

	assert.NotNil(t, e)
	assert.Equal(t, e.Exchange, "TestUnitExchange")
	assert.Equal(t, e.RoutingKey, "TestUnitKey")
	assert.Equal(t, e.DeliveryMode, amqp.Persistent)
	assert.Equal(t, e.Mandatory, false)
	assert.Equal(t, e.Ctx, context.Background())
	assert.NotNil(t, e.Headers)
	assert.NotEmpty(t, e.Headers)
	assert.Equal(t, e.Headers["TestUnitHeader"], "TestUnitValue")
}

func TestLetterCreate(t *testing.T) {
	e := fr.NewEnvelope(context.Background(), "Exchange", "Key", nil)
	id := uuid.New()
	l := fr.NewLetter(id, e, []byte("TestUnitQueue"))

	assert.NotNil(t, l)
	assert.Equal(t, l.LetterID, id)
	assert.NotNil(t, l.Envelope)
	assert.Equal(t, e, l.Envelope)
	assert.Equal(t, []byte("TestUnitQueue"), l.Body)
	assert.Equal(t, uint32(3), l.RetryCount)
}

func TestTableNewLetterWithPayload(t *testing.T) {
	e := fr.NewEnvelope(context.Background(), "Exchange", "Key", nil)
	hashy := crypto.GetHashWithArgon("password", "salt", 1, 12, 64, 32)

	encrypt := &fr.EncryptionConfig{
		Enabled:           false,
		Hashkey:           hashy,
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

			l, err := fr.NewLetterWithPayload(e, data, encryption, compression, false, "")

			assert.NotNil(t, l)
			assert.NoError(t, err)

			assert.NotNil(t, l.LetterID)
			assert.NotEmpty(t, l.LetterID)
			assert.NotEqual(t, uuid.Nil, l.LetterID)

			assert.Equal(t, uint32(3), l.RetryCount)

			assert.NotNil(t, l.Envelope)
			assert.Equal(t, e, l.Envelope)

			assert.NoError(t, err)
			assert.Equal(t, expectedData, l.Body)
		})
	}
}

package friendlyrabbit

import (
	"bytes"
	"os"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/pmatteo/friendlyrabbit/internal/utils/compression"
	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
	utils_json "github.com/pmatteo/friendlyrabbit/internal/utils/json"
)

var json = jsoniter.ConfigFastest

func ConvertJsonToStruct[C interface{}](fileNamePath string, c *C) error {
	byteValue, err := os.ReadFile(fileNamePath)
	if err != nil {
		return nil
	}

	return json.Unmarshal(byteValue, c)
}

// ConvertJSONFileToConfig opens a file.json and converts to RabbitSeasoning.
func ConvertJSONFileToConfig(fileNamePath string) (*RabbitSeasoning, error) {
	config := &RabbitSeasoning{}
	err := ConvertJsonToStruct(fileNamePath, config)

	return config, err
}

// ConvertJSONFileToTopologyConfig opens a file.json and converts to Topology.
func ConvertJSONFileToTopologyConfig(fileNamePath string) (*TopologyConfig, error) {

	config := &TopologyConfig{}
	err := ConvertJsonToStruct(fileNamePath, config)

	return config, err
}

// CreatePayload creates a JSON marshal and optionally compresses and encrypts the bytes.
func CreatePayload(
	input interface{},
	compressionConf *CompressionConfig,
	encryptionConf *EncryptionConfig,
) ([]byte, error) {

	data, err := json.Marshal(&input)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	if compressionConf != nil && compressionConf.Enabled {
		err := compression.Compress(compressionConf.Type, data, buffer)
		if err != nil {
			return nil, err
		}

		// Update data - data is now compressed
		data = buffer.Bytes()
	}

	if encryptionConf != nil && encryptionConf.Enabled {
		err := crypto.Encrypt(encryptionConf.Type, encryptionConf.Hashkey, data, buffer)
		if err != nil {
			return nil, err
		}

		// Update data - data is now encrypted
		data = buffer.Bytes()
	}

	return data, nil
}

// CreateWrappedPayload wraps your data in a plaintext wrapper called ModdedLetter and performs the selected modifications to data.
func CreateWrappedPayload(
	input interface{},
	letterID uuid.UUID,
	metadata string,
	compressionConf *CompressionConfig,
	encryptionConf *EncryptionConfig,
) ([]byte, error) {

	wrappedBody := &WrappedBody{
		LetterID:       letterID,
		LetterMetadata: metadata,
		Body:           &ModdedBody{},
	}

	var err error
	var innerData []byte
	innerData, err = json.Marshal(&input)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	if compressionConf != nil && compressionConf.Enabled {
		err := compression.Compress(compressionConf.Type, innerData, buffer)
		if err != nil {
			return nil, err
		}

		// Data is now compressed
		wrappedBody.Body.Compressed = true
		wrappedBody.Body.CType = compressionConf.Type
		innerData = buffer.Bytes()
	}

	if encryptionConf != nil && encryptionConf.Enabled {
		err := crypto.Encrypt(encryptionConf.Type, encryptionConf.Hashkey, innerData, buffer)
		if err != nil {
			return nil, err
		}

		// Data is now encrypted
		wrappedBody.Body.Encrypted = true
		wrappedBody.Body.EType = encryptionConf.Type
		innerData = buffer.Bytes()
	}

	wrappedBody.Body.UTCDateTime = utils_json.UtcTimestamp()
	wrappedBody.Body.Data = innerData

	data, err := json.Marshal(&wrappedBody)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ReadPayload unencrypts and uncompresses payloads
func ReadPayload(
	buffer *bytes.Buffer,
	compressionConf *CompressionConfig,
	encryptionConf *EncryptionConfig,
) error {

	if encryptionConf != nil && encryptionConf.Enabled {
		if err := crypto.Decrypt(encryptionConf.Type, encryptionConf.Hashkey, buffer); err != nil {
			return err
		}
	}

	if compressionConf != nil && compressionConf.Enabled {
		if err := compression.Decompress(compressionConf.Type, buffer); err != nil {
			return err
		}
	}

	return nil
}

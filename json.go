package friendlyrabbit

import (
	"bytes"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/pmatteo/friendlyrabbit/internal/utils/compression"
	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
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

// ToPayload creates a JSON marshal and optionally compresses and encrypts the bytes.
func ToPayload(
	input interface{},
	compressionConf *CompressionConfig,
	encryptionConf *EncryptionConfig,
) ([]byte, error) {

	var data []byte
	var err error
	var ok bool

	data, ok = input.([]byte)
	if !ok {
		data, err = json.Marshal(&input)
		if err != nil {
			return nil, err
		}
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

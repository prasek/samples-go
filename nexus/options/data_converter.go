package options

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

/*

const (
	// MetadataEncodingEncrypted is "binary/encrypted"
	MetadataEncodingEncrypted = "binary/encrypted"

	// MetadataEncryptionKeyID is "encryption-key-id"
	MetadataEncryptionKeyID = "encryption-key-id"
)

type DataConverter struct {
	// Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.
	parent converter.DataConverter
	converter.DataConverter
	options DataConverterOptions
}

type DataConverterOptions struct {
	KeyID string
	// Enable ZLib compression before encryption.
	Compress bool
}

// Codec implements PayloadCodec using AES Crypt.
type Codec struct {
	KeyID string
}

// TODO: Implement workflow.ContextAware in CodecDataConverter
// Note that you only need to implement this function if you need to vary the encryption KeyID per workflow.
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyID

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
// Note that you only need to implement this function if you need to vary the encryption KeyID per workflow.
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyID

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

func (e *Codec) getKey(keyID string) (key []byte) {
	// Key must be fetched from secure storage in production (such as a KMS).
	// For testing here we just hard code a key.
	return []byte("test-key-test-key-test-key-test!")
}

// NewEncryptionDataConverter creates a new instance of EncryptionDataConverter wrapping a DataConverter
func NewEncryptionDataConverter(dataConverter converter.DataConverter, options DataConverterOptions) *DataConverter {
	codecs := []converter.PayloadCodec{
		&Codec{KeyID: options.KeyID},
	}
	// Enable compression if requested.
	// Note that this must be done before encryption to provide any value. Encrypted data should by design not compress very well.
	// This means the compression codec must come after the encryption codec here as codecs are applied last -> first.
	if options.Compress {
		codecs = append(codecs, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))
	}

	return &DataConverter{
		parent:        dataConverter,
		DataConverter: converter.NewCodecDataConverter(dataConverter, codecs...),
		options:       options,
	}
}
*/

//TODO: Merge above WithWorkflowContext WithContext for ContextAware support specifically for Nexus contexts
//      so we can no encrypt select Nexus payloads
//      alternatively or additively we can choose to not encrypt only nexus.NoValue == nil (which is what we do below)

const (
	// metadataEncodingEncrypted identifies payloads encoded with an encrypted binary format
	metadataEncodingEncrypted = "binary/encrypted"
	// metadataEncryptionKeyID identifies the key used to encrypt a payload
	metadataEncryptionKeyID = "encryption-key-id"
)

// Codec provides methods for encrypting and decrypting payload data
type Codec struct {
	EncryptionKeyID string
}

// this function simulates the retrieval of an encryption key (identified by
// the provided key ID) from secure storage, such as a key management server
func (e *Codec) retrieveKey(keyID string) (key []byte, err error) {
	if keyID == "" {
		return nil, fmt.Errorf("key retrieval failed due to empty identifier")
	}

	// Simulate key retrieval by using a hash function to generate
	// a 256-bit value that will be consistent for a given key ID
	h := sha256.Sum256([]byte(keyID))

	return h[:], nil
}

// NewEncryptionDataConverter creates and returns a DataConverter instance that
// wraps the default DataConverter with a CodecDataConverter that uses encryption
// to protect the confidentiality of payload data. This instance will encrypt data
// using a key associated with the specified encryption key ID.
func NewEncryptionDataConverter(underlying converter.DataConverter, encryptionKeyID string) converter.DataConverter {
	codecs := []converter.PayloadCodec{
		&Codec{EncryptionKeyID: encryptionKeyID},
	}

	return converter.NewCodecDataConverter(underlying, codecs...)
}

// Encode implements the Encode method defined by the converter.PayloadCodec interface
func (e *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {

		fmt.Printf("Encode Payload.Data: len(%d) %#v\n", len(payload.Data), payload.Data)

		payloadFormatID := string(payload.Metadata[converter.MetadataEncoding])
		fmt.Printf("Encode Payload.Encoding: len(%d) %#v\n", len(payloadFormatID), payloadFormatID)

		// don't encrypt nexus no value if not using same KMS keyID or passing key ID in payload metadata
		if payloadFormatID == MetadataEncodingNexusNoValue {
			result[i] = payload
			continue
		}

		//encrypt everything else
		unencryptedData, err := payload.Marshal()
		if err != nil {
			return payloads, err
		}

		fmt.Printf("Encode Unencrypted Data: len(%d) %#v\n", len(unencryptedData), unencryptedData)

		key, err := e.retrieveKey(e.EncryptionKeyID)
		if err != nil {
			return payloads, err
		}

		encryptedData, err := encrypt(unencryptedData, key)
		if err != nil {
			return payloads, err
		}

		fmt.Printf("Encode Encrypted Data: len(%d) %#v\n", len(encryptedData), encryptedData)

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte(metadataEncodingEncrypted),
				metadataEncryptionKeyID:    []byte(e.EncryptionKeyID),
			},
			Data: encryptedData,
		}

		fmt.Printf("Encoded EncryptionKeyID (on wire): %#v\n", string(result[i].Metadata[metadataEncryptionKeyID]))
	}

	return result, nil
}

// Decode implements the Decode method defined by the converter.PayloadCodec interface
func (e *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		payloadFormatID := string(payload.Metadata[converter.MetadataEncoding])
		fmt.Printf("Decode Payload.Encoding: len(%d) %#v\n", len(payloadFormatID), payloadFormatID)

		// Skip decryption for any payload not using our encrypted format
		if payloadFormatID != metadataEncodingEncrypted {
			result[i] = payload
			continue
		}

		encryptedData := payload.Data
		fmt.Printf("Decode encryptedData: len(%d) %#v\n", len(encryptedData), encryptedData)

		/*
			//don't use key sent on wire, only use my key to simulate separate keys per ns/region
			fmt.Printf("Using EncryptionKeyID (from local ns/region): %#v\n", string(e.EncryptionKeyID))
			key, err := e.retrieveKey(e.EncryptionKeyID)
		*/

		keyID, ok := payload.Metadata[metadataEncryptionKeyID]
		if !ok {
			return payloads, fmt.Errorf("encryption key id missing from metadata")
		}
		fmt.Printf("Recieved EncryptionKeyID (on wire): %#v\n", string(keyID))
		key, err := e.retrieveKey(string(keyID))
		if err != nil {
			return payloads, err
		}

		decryptedData, err := decrypt(encryptedData, key)
		if err != nil {
			return payloads, err
		}

		fmt.Printf("Decode decryptedData: len(%d) %#v\n", len(decryptedData), decryptedData)

		result[i] = &commonpb.Payload{}
		err = result[i].Unmarshal(decryptedData)
		if err != nil {
			fmt.Printf("Unmarshal Error: %#v\n", err)
			return payloads, err
		}
		fmt.Printf("Decode Payload.Data: len(%d) %#v\n", len(result[i].Data), result[i].Data)
	}

	return result, nil
}

// Uses AES to encrypt the provided block of data using with the provided key
func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := aesgcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// Uses AES to decrypt the provided block of data using the provided key
func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	return aesgcm.Open(nil, nonce, encrypted, nil)
}

/*

func isInterfaceNil(i interface{}) bool {
	if nexusNoValue, ok := i.(nexus.NoValue); ok {
		nexusNoValueIsNil := nexusNoValue == nil
		fmt.Printf("nexusNoValueIsNil %#v\n", nexusNoValueIsNil)
		return nexusNoValueIsNil
	}
	fmt.Println("isInterfaceNil: FALSE")
	return false

	fmt.Printf("isInterfaceNil %#v\n", i)
	v := reflect.ValueOf(i)
	return i == nil || ((v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice) && v.IsNil())
}

*/

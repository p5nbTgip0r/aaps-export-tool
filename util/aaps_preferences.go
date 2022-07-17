package util

import (
	"encoding/hex"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

const (
	// KeyConscience is declared here:
	// https://github.com/nightscout/AndroidAPS/blob/219bdba21531c8f9d5df0ebaf5a7a3821c179d9a/core/src/main/java/info/nightscout/androidaps/plugins/general/maintenance/formats/EncryptedPrefsFormat.kt#L30
	KeyConscience = "if you remove/change this, please make sure you know the consequences!"

	FileHashPlaceholder = "--to-be-calculated--"
)

func CalculateFileHash(exportJson []byte) []byte {
	exportJson, _ = sjson.SetBytes(exportJson, "security.file_hash", FileHashPlaceholder)

	hash := Hmac256(exportJson, KeyConscience)
	exportJson, _ = sjson.SetBytes(exportJson, "security.file_hash", hash)
	return exportJson
}

func IsEncrypted(exportJson []byte) bool {
	return gjson.GetBytes(exportJson, "format").String() == "aaps_encrypted"
}

func IsPreferencesObject(exportJson []byte) bool {
	return gjson.GetBytes(exportJson, "content").IsObject()
}

func ConvertPreferencesToObject(exportJson []byte) []byte {
	content := gjson.GetBytes(exportJson, "content").String()
	export, _ := sjson.SetRawBytes(exportJson, "content", []byte(content))
	export = pretty.Pretty(export)
	return CalculateFileHash(export)
}

func ConvertPreferencesToString(exportJson []byte) []byte {
	content := gjson.GetBytes(exportJson, "content|@ugly").Raw
	export, _ := sjson.SetBytes(exportJson, "content", content)
	export = pretty.Pretty(export)
	return CalculateFileHash(export)
}

func ConvertToUnencryptedFormat(encryptedExportJson []byte, decryptedContent []byte) []byte {
	var output = encryptedExportJson
	output, _ = sjson.SetBytes(output, "format", "aaps_structured")
	output, _ = sjson.SetBytes(output, "security.file_hash", FileHashPlaceholder)
	output, _ = sjson.SetBytes(output, "security.algorithm", "none")
	output, _ = sjson.SetBytes(output, "content", string(decryptedContent))

	output, _ = sjson.DeleteBytes(output, "security.salt")
	output, _ = sjson.DeleteBytes(output, "security.content_hash")

	return CalculateFileHash(output)
}

func ConvertToEncryptedFormat(unencryptedExportJson []byte, salt []byte, encryptedContent []byte, contentHash string) []byte {
	var output = unencryptedExportJson
	output, _ = sjson.SetBytes(output, "format", "aaps_encrypted")
	output, _ = sjson.SetBytes(output, "security.file_hash", FileHashPlaceholder)
	output, _ = sjson.SetBytes(output, "security.algorithm", "v1")
	output, _ = sjson.SetBytes(output, "security.salt", hex.EncodeToString(salt))
	output, _ = sjson.SetBytes(output, "security.content_hash", contentHash)
	output, _ = sjson.SetBytes(output, "content", string(encryptedContent))
	output = pretty.Pretty(output)

	return CalculateFileHash(output)
}

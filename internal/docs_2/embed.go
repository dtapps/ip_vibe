package docs_2

import _ "embed"

//go:embed swagger.json
var swaggerJSONBytes []byte

func GetSwaggerBytes() []byte {
	return swaggerJSONBytes
}

module vibe_ip

go 1.25.5

ignore ./cmd/backend/static

ignore ./cmd/backend/templates

ignore ./cmd/frontend

ignore ./sql

replace github.com/quic-go/qpack => github.com/quic-go/qpack v0.6.0

replace github.com/quic-go/quic-go => github.com/quic-go/quic-go v0.58.0

replace github.com/imroc/req/v3 => github.com/imroc/req/v3 v3.57.0

require (
	entgo.io/ent v0.14.5
	github.com/alibabacloud-go/alidns-20150109/v5 v5.1.1
	github.com/alibabacloud-go/cdn-20180510/v9 v9.0.0
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.14
	github.com/alibabacloud-go/esa-20240910/v2 v2.44.0
	github.com/alibabacloud-go/tea v1.4.0
	github.com/aliyun/credentials-go v1.4.10
	github.com/apache/thrift v0.22.0
	github.com/baidubce/bce-sdk-go v0.9.258
	github.com/cloudflare/cloudflare-go/v6 v6.6.0
	github.com/corazawaf/coraza-coreruleset v0.0.0-20240226094324-415b1017abdc
	github.com/corazawaf/coraza/v3 v3.3.3
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dromara/carbon/v2 v2.6.15
	github.com/envoyproxy/protoc-gen-validate v1.3.0
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-contrib/gzip v1.2.5
	github.com/gin-contrib/multitemplate v1.1.1
	github.com/gin-contrib/requestid v1.0.5
	github.com/gin-gonic/gin v1.11.0
	github.com/go-acme/lego/v4 v4.31.0
	github.com/go-co-op/gocron/v2 v2.19.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.4
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.184
	github.com/imroc/req/v3 v3.57.0
	github.com/jdcloud-api/jdcloud-sdk-go v1.64.0
	github.com/kingsoftcloud/sdk-go/v2 v2.2.59
	github.com/miekg/dns v1.1.70
	github.com/modelcontextprotocol/go-sdk v1.2.0
	github.com/qiniu/go-sdk/v7 v7.25.6
	github.com/spf13/cast v1.10.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	github.com/swaggo/swag v1.16.6
	github.com/swaggo/swag/v2 v2.0.0-rc5
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn v1.3.22
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.33
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.3.24
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo v1.3.32
	github.com/volcengine/volcengine-go-sdk v1.2.7
	github.com/wdcbot/qingfeng v1.6.3
	go.dtapp.net/library/contrib/http_log v1.0.9
	go.dtapp.net/library/contrib/req_log v1.0.3
	go.dtapp.net/library/contrib/resty_log v1.0.14
	go.dtapp.net/library/service/bark v1.0.9
	go.dtapp.net/library/service/dingtalk v1.0.41
	go.dtapp.net/library/service/feishu v1.0.39
	go.dtapp.net/library/service/ft07 v1.0.10
	go.dtapp.net/library/service/pushdeer v1.0.10
	go.dtapp.net/library/service/wechatqy v1.0.40
	go.dtapp.net/library/utils/goarray v1.0.4
	go.dtapp.net/library/utils/gojson v1.0.13
	go.dtapp.net/library/utils/golog v1.0.180
	go.dtapp.net/library/utils/gomd5 v1.0.3
	go.dtapp.net/library/utils/gorandom v1.0.5
	go.dtapp.net/library/utils/gorequest v1.1.3
	go.dtapp.net/library/utils/gostring v1.0.32
	go.dtapp.net/library/utils/gotime v1.0.17
	golang.org/x/net v0.49.0
	golang.org/x/sync v0.19.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260114163908-3f89685c29c3
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260114163908-3f89685c29c3
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	modernc.org/sqlite v1.44.1
	resty.dev/v3 v3.0.0-beta.6
)

require (
	ariga.io/atlas v1.0.0 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/MercuryEngineering/CookieMonster v0.0.0-20180304172713-1584578b3403 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/basgys/goxml2json v1.1.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.2 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/corazawaf/libinjection-go v0.2.3 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-acme/alidns-20150109/v4 v4.7.0 // indirect
	github.com/go-acme/esa-20240910/v2 v2.44.0 // indirect
	github.com/go-acme/tencentclouddnspod v1.1.25 // indirect
	github.com/go-acme/tencentedgdeone v1.1.48 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-openapi/inflect v0.21.5 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/spec v0.22.3 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/google/pprof v0.0.0-20251213031049-b05bdaca462f // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/icholy/digest v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magefile/mage v1.15.1-0.20241126214340-bdc92f694516 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/nrdcg/dnspod-go v0.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20250424160509-463d218d4745 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	github.com/refraction-networking/utls v1.8.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sv-tools/openapi v0.4.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.235 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/zclconf/go-cty v1.17.0 // indirect
	github.com/zclconf/go-cty-yaml v1.2.0 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.23.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	gopkg.in/ini.v1 v1.67.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)

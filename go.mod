module vibe_ip

go 1.27.0

ignore ./cmd/backend/static

ignore ./cmd/backend/templates

ignore ./cmd/frontend

ignore ./sql

require (
	entgo.io/ent v0.14.6
	github.com/alibabacloud-go/alidns-20150109/v5 v5.6.0
	github.com/alibabacloud-go/cdn-20180510/v9 v9.0.0
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.2.4
	github.com/alibabacloud-go/esa-20240910/v2 v2.48.0
	github.com/alibabacloud-go/tea v1.5.3
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9
	github.com/aliyun/credentials-go v1.4.13
	github.com/baidubce/bce-sdk-go v0.9.274
	github.com/chenyahui/gin-cache v1.10.0
	github.com/cloudflare/cloudflare-go/v6 v6.10.0
	github.com/corazawaf/coraza-coreruleset v0.0.0-20240226094324-415b1017abdc
	github.com/corazawaf/coraza/v3 v3.7.0
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dromara/carbon/v2 v2.6.17
	github.com/dustin/go-humanize v1.0.1
	github.com/envoyproxy/protoc-gen-validate v1.3.3
	github.com/gin-contrib/cors v1.7.7
	github.com/gin-contrib/gzip v1.2.6
	github.com/gin-contrib/multitemplate v1.1.2
	github.com/gin-contrib/requestid v1.0.6
	github.com/gin-gonic/gin v1.12.0
	github.com/go-acme/lego/v5 v5.3.1
	github.com/go-co-op/gocron/v2 v2.22.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.212
	github.com/inetaf/tcpproxy v0.0.0-20260515195445-c159a6051109
	github.com/jdcloud-api/jdcloud-sdk-go v1.67.0
	github.com/kingsoftcloud/sdk-go/v2 v2.2.122
	github.com/miekg/dns v1.1.73
	github.com/modelcontextprotocol/go-sdk v1.7.0
	github.com/qiniu/go-sdk/v7 v7.27.0
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/spf13/cast v1.10.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	github.com/swaggo/swag v1.16.6
	github.com/swaggo/swag/v2 v2.0.0-rc5
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn v1.3.154
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.165
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.3.131
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo v1.3.158
	github.com/tidwall/gjson v1.19.0
	github.com/volcengine/volcengine-go-sdk v1.2.48
	github.com/wdcbot/qingfeng v1.6.8
	go.dtapp.net/library/contrib/http_log v1.0.13
	go.dtapp.net/library/contrib/resty_log v1.0.18
	go.dtapp.net/library/service/bark v1.0.11
	go.dtapp.net/library/service/dingtalk v1.0.43
	go.dtapp.net/library/service/feishu v1.0.41
	go.dtapp.net/library/service/ft07 v1.0.12
	go.dtapp.net/library/service/pushdeer v1.0.12
	go.dtapp.net/library/service/qiniu v1.0.0
	go.dtapp.net/library/service/wechatqy v1.0.42
	go.dtapp.net/library/utils/gojson v1.0.13
	go.dtapp.net/library/utils/golog v1.0.180
	go.dtapp.net/library/utils/gomd5 v1.0.3
	go.dtapp.net/library/utils/gorandom v1.0.5
	go.dtapp.net/library/utils/gorequest v1.1.3
	go.dtapp.net/library/utils/gostring v1.0.32
	go.dtapp.net/library/utils/gotime v1.0.19
	golang.org/x/net v0.58.0
	golang.org/x/sync v0.22.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260819154853-08b0e4226688
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688
	google.golang.org/grpc v1.83.1
	google.golang.org/protobuf v1.36.12
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	modernc.org/sqlite v1.57.0
	resty.dev/v3 v3.0.0-rc.3
)

require (
	ariga.io/atlas v1.3.0 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/MercuryEngineering/CookieMonster v0.0.0-20180304172713-1584578b3403 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/apparentlymart/go-textseg/v17 v17.0.1 // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/basgys/goxml2json v1.1.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.2 // indirect
	github.com/bytedance/sonic/loader v0.5.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/corazawaf/libinjection-go v0.3.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/gin-contrib/sse v1.1.1 // indirect
	github.com/go-acme/alidns-20150109/v5 v5.6.0 // indirect
	github.com/go-acme/esa-20240910/v3 v3.13.0 // indirect
	github.com/go-acme/tencentclouddnspod v1.3.24 // indirect
	github.com/go-acme/tencentedgdeone v1.3.38 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-json-experiment/json v0.0.0-20260820222146-c27c302e5fc3 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/inflect v1.0.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-openapi/jsonreference v1.0.0 // indirect
	github.com/go-openapi/spec v0.22.9 // indirect
	github.com/go-openapi/swag/conv v0.29.0 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.0 // indirect
	github.com/go-openapi/swag/loading v0.29.0 // indirect
	github.com/go-openapi/swag/pools v0.29.0 // indirect
	github.com/go-openapi/swag/stringutils v0.29.0 // indirect
	github.com/go-openapi/swag/typeutils v0.29.0 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/jellydator/ttlcache/v2 v2.11.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/kaptinlin/jsonpointer v0.4.28 // indirect
	github.com/kaptinlin/jsonschema v0.9.8 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/onsi/gomega v1.37.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20250424160509-463d218d4745 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20260805114148-88456608a4f6 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.61.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/shoenig/go-m1cpu v0.2.2 // indirect
	github.com/sv-tools/openapi v0.4.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.2 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.255 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zclconf/go-cty v1.19.0 // indirect
	github.com/zclconf/go-cty-yaml v1.2.0 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	go.mongodb.org/mongo-driver/v2 v2.8.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/arch v0.30.0 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.75.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
	software.sslmate.com/src/go-pkcs12 v0.7.3 // indirect
)

replace github.com/wdcbot/qingfeng => gitee.com/xiaowan1997/qingfeng v1.6.7

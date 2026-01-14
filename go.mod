module novel-video-workflow

go 1.25.5

require (
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jessevdk/go-flags v1.6.1
	github.com/mark3labs/mcp-go v0.43.2
	github.com/spf13/viper v1.21.0
	go.uber.org/zap v1.27.1
	golang.org/x/image v0.34.0
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace novel-video-workflow/pkg/mcp => ./pkg/mcp

replace novel-video-workflow/pkg/workflow => ./pkg/workflow

replace novel-video-workflow/mcp_pkg => ./pkg/mcp

replace novel-video-workflow/pkg/tools/drawthings => ./pkg/tools/drawthings

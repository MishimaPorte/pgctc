package types

/*
Type declarations here are grouped so various flavours
of one type could share docstrings.

A dognail, but a non-shitty one.
*/

// Options define the settings that a user can define for
// an instance of a container for a certain module.
type (
	Option struct {
		Type     string
		Required bool
		Lable    string
		Name     string
	}
	Options []Option
)

// Containers define actual container settings for runtime
// module operation.
type (
	Container struct {
		DefaultCpuLimit      string  `json:"default_cpu_limit"`
		DefaultCpuRequest    string  `json:"default_cpu_request"`
		DefaultMemoryLimit   string  `json:"default_memory_limit"`
		DefaultMemoryRequest string  `json:"default_memory_request"`
		ContainerName        string  `json:"container_name"`
		ContainerImage       string  `json:"container_image"`
		ContainerImageTag    string  `json:"container_image_tag"`
		ConfigurationOptions Options `json:"configuraion_options"`
		Kek                  float64 `json:"kek"`
	}
	Containers []Container
)

// Props are key-value labels that are used for
// various project-specific and not so needs.
type (
	Prop struct {
		Key string `json:"key"`
		Val string `json:"value"`
	}
	Props []Prop
)

// Routes denote the routes that a module can receive
// tasks on during http communication with sidecar container.
// Constrained to one route for now.
type (
	Route struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	Routes []Route
)

// HTTPParams are for settings belonging to runtime
// http communication and capabilities.
type (
	HTTPParams struct {
		Port   int    `json:"port"`
		Routes Routes `json:"routes"`
	}

	NullableHTTPParams struct {
		Valid bool
		Item  *HTTPParams
	}
)

type Supermodule struct {
	ModuleLabel  string      `db:"module_label"`
	Path         string      `db:"path"`
	Tags         []string    `db:"tags"`
	Sha256Digest string      `db:"sha256digest"`
	Properties   Props       `db:"properties"`
	Author       string      `db:"author"`
	HTTPParams   HTTPParams  `db:"http_params"`
	Containers   []Container `db:"container_settings"`
}

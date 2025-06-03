package types

import "database/sql/driver"

/*
Type declarations here are grouped so various flavours
of one type could share docstrings.

A dognail, but a non-shitty one.
*/
type (
	//#[generate(Scanner)]
	Things []string
	//#[generate(Scanner)]
	//#[generate(Valuer)]
	Option struct {
		Type     string
		A        int
		B        int16
		C        uint
		D        ******uintptr
		Keks     []bool
		Things2  []string
		Required bool
		Lable    lol
		Name     []kek
		Auf      A
		Array    [10]bool
		Slice    []bool
		BadThing struct {
			Kek string
			Lol []struct {
				Kek string
				Lol bool
			}
		}
		Hash ShaDigest
	}
)

func (s ShaDigest) Scan(thing any) (err error) {
	return nil
}
func (s ShaDigest) Value() (v driver.Value, err error) {
	return nil, nil
}

type ShaDigest [32]byte

type A struct {
	Kek string
}

type kek struct {
	Kek string
}
type lol kek

//
// // Containers define actual container settings for runtime
// // module operation.
// type (
// 	Container struct {
// 		DefaultCpuLimit      string   `json:"default_cpu_limit"`
// 		DefaultCpuRequest    string   `json:"default_cpu_request"`
// 		DefaultMemoryLimit   string   `json:"default_memory_limit"`
// 		DefaultMemoryRequest string   `json:"default_memory_request"`
// 		ContainerName        string   `json:"container_name"`
// 		ContainerImage       string   `json:"container_image"`
// 		ContainerImageTag    string   `json:"container_image_tag"`
// 		ConfigurationOptions []Option `json:"configuraion_options"`
// 		Kek                  float64  `json:"kek"`
// 	}
// 	Containers []Container
// )
//
// type KekValue string
//
// const (
// 	Kek KekValue = "kek"
// 	Lol KekValue = "lol"
// )
//
// // Props are key-value labels that are used for
// // various project-specific and not so needs.
// type (
// 	Prop struct {
// 		Key string `json:"key"`
// 		Val string `json:"value"`
// 	}
// 	Props []Prop
// )
//
// // Routes denote the routes that a module can receive
// // tasks on during http communication with sidecar container.
// // Constrained to one route for now.
// type (
// 	Route struct {
// 		Path   string         `json:"path"`
// 		Method sql.NullString `json:"method"`
// 	}
// 	Routes []Route
// )
//
// // HTTPParams are for settings belonging to runtime
// // http communication and capabilities.
// type (
// 	HTTPParams struct {
// 		Port   int    `json:"port"`
// 		Routes Routes `json:"routes"`
// 	}
//
// 	NullableHTTPParams struct {
// 		Valid bool
// 		Item  *HTTPParams
// 	}
// )
//
// type Supermodule struct {
// 	ModuleLabel  string      `db:"module_label"`
// 	Path         string      `db:"path"`
// 	Tags         []string    `db:"tags"`
// 	Sha256Digest string      `db:"sha256digest"`
// 	Properties   Props       `db:"properties"`
// 	Author       string      `db:"author"`
// 	HTTPParams   HTTPParams  `db:"http_params"`
// 	Containers   []Container `db:"container_settings"`
// }

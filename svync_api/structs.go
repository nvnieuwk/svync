package svync_api

// VCF structs
type VCF struct {
	Header   Header
	Variants map[string]Variant
}

type Header struct {
	Info    map[string]HeaderLineIdNumberTypeDescription
	Format  map[string]HeaderLineIdNumberTypeDescription
	Alt     map[string]HeaderLineIdDescription
	Filter  map[string]HeaderLineIdDescription
	Contig  map[string]HeaderLineIdLength
	Other   []string
	Samples []string
}

type HeaderLineIdDescription struct {
	Id          string
	Description string
}

type HeaderLineIdNumberTypeDescription struct {
	Id          string
	Number      string
	Type        string
	Description string
}

type HeaderLineIdLength struct {
	Id     string
	Length int64
}

type Variant struct {
	Chromosome string
	Pos        int64
	Id         string
	Ref        string
	Alt        string
	Qual       string
	Filter     string
	Header     *Header
	Info       map[string][]string
	Format     map[string]VariantFormat
	Parsed     bool
}

type VariantFormat struct {
	Sample  string
	Content map[string][]string
}

//
// Config structs
//

type Config struct {
	Id     string
	Alt    map[string]string
	Info   MapConfigInput
	Format MapConfigInput
}

type MapConfigInput map[string]ConfigInput
type ConfigInput struct {
	Value       string
	Description string
	Number      string
	Type        string
	Alts        map[string]string
}

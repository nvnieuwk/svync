package svync_api

type VCF struct {
	Header   Header
	Variants map[string]VariantMates
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

type VariantMates struct {
	Mate1 Variant
	Mate2 Variant
}

type Variant struct {
	Chromosome string
	Start      int64
	End        int64
	Ref        string
	Alt        string
	Qual       string
	Filter     string
	Info       map[string]string
	Format     map[string]VariantFormat
}

type VariantFormat struct {
	Sample string
	Format map[string]string
}

package svync_api

// The struct representing the header of the input VCF file in a parseable format
type Header struct {
	// Object containing the INFO fields with their ID, Number, Type and Description
	// The ID is the key of the map
	// The value is a struct containing the Id, Number, Type and Description
	Info map[string]HeaderLineIdNumberTypeDescription

	// Object containing the FORMAT fields with their ID, Number, Type and Description
	// The ID is the key of the map
	// The value is a struct containing the Id, Number, Type and Description
	Format map[string]HeaderLineIdNumberTypeDescription

	// Object containing the ALT fields with their ID and Description
	// The ID is the key of the map
	// The value is a struct containing the Id and Description
	Alt map[string]HeaderLineIdDescription

	// Object containing the FILTER fields with their ID and Description
	// The ID is the key of the map
	// The value is a struct containing the Id and Description
	Filter map[string]HeaderLineIdDescription

	// List of all contigs in the VCF file with their ID and Length
	Contig []HeaderLineIdLength

	// List of all other VCF fields
	Other []string

	// List of all samples in the VCF file
	Samples []string
}

// A struct representing a header line in the VCF file with its ID and Description
type HeaderLineIdDescription struct {
	// The ID of the header line
	Id string

	// The description of the header line
	Description string
}

// A struct representing a header line in the VCF file with its ID, Number, Type and Description
type HeaderLineIdNumberTypeDescription struct {
	// The ID of the header line
	Id string

	// The number of values in the header line
	// Can be any integer, "A", "G", "R" or "."
	// A = one value per alternate allele
	// G = one value per possible genotype
	// R = one value per possible allele
	// . = the number varies, is unkown or is unbounded
	Number string

	// The type of the header line
	// Can be "Integer", "Float", "Flag", "String" or "Character"
	Type string

	// The description of the header line
	Description string
}

// A struct representing a header line in the VCF file with its ID and Length
type HeaderLineIdLength struct {
	// The ID of the header line
	Id string

	// The length of the header line
	Length int64
}

// A struct representing a variant in the input VCF file
type Variant struct {
	// The chromosome of the variant
	Chromosome string

	// The 1-based position of the variant
	Pos int64

	// The ID of the variant
	Id string

	// The reference allele of the variant
	Ref string

	// The alternate allele of the variant
	Alt string

	// The Phred-scaled quality score of the variant
	Qual string

	// The filter status of the variant
	Filter string

	// A pointer to the header of the VCF that contains this variant
	Header *Header

	// The INFO values of the variant
	Info map[string][]string

	// The FORMAT values of the variant
	Format map[string]VariantFormat

	// A status flag indicating if the variant has been parsed before
	Parsed bool
}

// A struct representing the format of a variant in the VCF file
type VariantFormat struct {
	// The sample name of the variant
	Sample string

	// The content of the format field
	Content map[string][]string
}

//
// Config structs
//

// The struct representing the configuration file
// The config file is a YAML file
type Config struct {
	// How to handle the ID field of each variant
	Id string

	// How to handle the ALT field of each variant
	// A value can be given for each SVTYPE
	Alt ConfigSimpleInput

	// How to handle the INFO fields of each variant
	Info MapConfigInput

	// How to handle the FORMAT fields of each variant
	Format MapConfigInput
}

// A struct representing a simple configuration of a field
type ConfigSimpleInput struct {
	// The value of the field
	// This can be a string or a reference to another field
	Value string

	// Alternative values for each SVTYPE
	Alts map[string]string
}

// A map construct for advanced configurations
type MapConfigInput map[string]ConfigInput

// A struct representing the configuration of advanced fields (like INFO and FORMAT)
type ConfigInput struct {
	// The value of the field
	// This can be a string or a reference to another field
	Value string

	// The default values of the field of all resolvable values in the Value field
	Defaults map[string]string

	// The description of the field
	// This is used to generate the VCF header
	Description string

	// The number of values in the field
	Number string

	// The type of the field
	Type string

	// Alternative values for each SVTYPE
	Alts map[string]string
}

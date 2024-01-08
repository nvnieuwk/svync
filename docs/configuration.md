# Configuration
The configuration file consists of 4 main parts:
1. `id` 
2. `alt`
3. `info`
4. `format`

## `id`
The `id` section is used to define the ID of the variant. The `id` section can be defined as follows:
```yaml
id: <id>
```
The value for the ID can be resolved (see [Resolvable fields](#resolvable-fields)). All IDs get a unique number appended to them to ensure that they are unique.

## `alt`
The `alt` section can be used to change the ALT field and SVTYPE info field for each variant. The `alt` section can be defined as follows:
```yaml
alt:
  <alt>: <new_alt>
```

For example you might want to change the `BND` ALT to `TRA` (for Delly for example):
```yaml
alt:
  BND: TRA
```

## `info`
The `info` section can be used to change the info fields for each variant. The `info` section can be defined as follows:
```yaml
info:
  <info_field>:
    value: <new_value>
    type: <new_type>
    description: <new_description>
    number: <new_number>
    alts:
      <alt>: <new_value>
      <alt>: <new_value>
```
### value
The `value` field can be used to change the default value of the info field. The value can be resolved (see [Resolvable fields](#resolvable-fields)).

### type
The `type` field can be used to set the type of the info field (This will be reflected in the header of the output VCF file).

### description
The `description` field can be used to set the description of the info field (This will be reflected in the header of the output VCF file).

### number
The `number` field can be used to set the number of the info field (This will be reflected in the header of the output VCF file).

### alts
The `alts` field can be used to set the value of the info field for a specific ALT. The value can be resolved (see [Resolvable fields](#resolvable-fields)).

For example when all `SVLEN` info fields are positive, you maybe want to change the field for all deletions to the negative length:
```yaml
info:
  SVLEN:
    value: $INFO/SVLEN
    type: Integer
    description: "Structural variant length"
    number: 1
    alts:
      DEL: -$INFO/SVLEN
```

## `format`
The `format` section can be used to change the format fields for each variant. The `format` section can be defined as follows:
```yaml
format:
  <format_field>:
    value: <new_value>
    type: <new_type>
    description: <new_description>
    number: <new_number>
    alts:
      <alt>: <new_value>
      <alt>: <new_value>
```

The format fields work the same as the info fields (see [Info](#info)). 

## Resolvable fields

Some fields can be resolved to a value. 

### Variables

A variable can be resolved appending a `$` to the field name. 

Following variables are available:
1. `$FORMAT/<format_field>` => This is only accesible for other format fields
    - An additional `/<number>` can be added to get a specific value in case of multiple values
2. `$INFO/<info_field>`
    - An additional `/<number>` can be added to get a specific value in case of multiple values
3. `$POS`
4. `$CHROM`
5. `$ALT`
6. `$QUAL`
7. `$FILTER`

For example `$INFO/SVLEN` will be resolved to the value of the `SVLEN` info field.

### Functions

Functions are very simple calculations that can be done on the values.

More functions can be added in the future. Please open an issue to request new functions.

#### `~sub`
The `~sub` function can be used to substract values from each other. The function can be used as follows:

```yaml
~sub:<value_start>,<value_to_substract>,<value_to_substract>,...
```

:warning: only integers and floats are supported for this function :warning:

#### `~sum`
The `~sum` function can be used to take the sum of all values. The function can be used as follows:

```yaml
~sum:<value_start>,<value_to_add>,<value_to_add>,...
```

:warning: only integers and floats are supported for this function :warning:


# bb

Pictographic programming language. Designed to make metadata injection and data entry super quick and easy!

Define your own types and the syntax for them! Then bb converts them to JSON. For example:

    a = { type: apple, >: isRed }  // this is a type definion
    
    a 4a a> a:foo  // this is data

Is converted to:

```json
[
    { "type": "apple" },
    { "type": "apple", "quantity": 4 },
    { "type": "apple", "isRed": true },
    { "type": "apple", "value": "foo" }
] 
```

Try it out in your browser with the [**bb playground**](https://mattsimmons1.github.io/bb/playground/)!

### Install

If you have [go](https://golang.org/doc/install) installed you can build the binary yourself and install with:

```bash
go get github.com/MattSimmons1/bb
```

Alternatively, download the binary from the [releases page](https://github.com/MattSimmons1/bb/releases) and save to a location on your PATH.

There is also a [Python library](https://pypi.org/project/bb-python/) for using bb within scripts.

```bash
pip install bb-python
```

See the [Python client docs](./client/python) for code examples.

### Usage

bb can be used from the command line. It takes a string from bb syntax and outputs JSON:

```shell-session
$ bb "hello world 123"  
["hello", "world", 123]
```

To read from a file:

```shell-session
$ cat my_data.bb.txt
M = { type: message } 
M"hello world" 
$ bb my_data.bb.txt
[{ "type": "message", "value": "hello world" }]
```

### Basic Syntax

| Syntax            | Usage                        | Result                  |
|-------------------|------------------------------|-------------------------|
| number            | 12                           | `12`                    | 
| string            | foo                          | `"foo"`                 | 
| safe string       | "foo" or \`foo`              | `"foo"`                 | 
| array             | 1 2 "foo"                    | `[1, 2, "foo"]`         |
| user defined type | ∆ = { }<br>∆                 | `{}`                    | 
| quantity          | ∆ = { }<br>3∆                | `{ "quantity": 3 }`     |
| numeric value     | ∆ = { }<br>∆5                | `{ "value": 5 }`        |
| string value      | ∆ = { }<br>∆"foo"            | `{ "value": "foo" }`    |
| string value      | ∆ = { }<br>∆\`foo`           | `{ "value": "foo" }`    |
| string value      | ∆ = { }<br>∆:foo             | `{ "value": "foo" }`    |
| numeric prop      | ∆ = { foo: 100 }<br>∆        | `{ "foo": 100 }`        |
| string prop       | ∆ = { foo: bar }<br>∆        | `{ "foo": "bar" }`      |
| modifier          | ∆ = { +: foo }<br>∆+         | `{ "foo": true }`       |
| modifier value    | ∆ = { +: foo }<br>∆+1        | `{ "foo": 1 }`          |
| repeated modifier | ∆ = { +: foo }<br>∆+3+\`bar` | `{ "foo": [3, "bar"] }` |
| script prop       | ∆ = { foo: d => 2 * 2 }<br>∆ | `{ "foo": 4 }`          |

### Reserved Characters, Keywords, and Other Syntax

These can't be used as units or modifiers

| Syntax             | Meaning                                                  |
|--------------------|----------------------------------------------------------|
| =                  | Defines a type                                           |
| .                  | Decimal point                                            |
| -                  | Negative sign                                            |
| { }                | Start and end of a code block or structure               |
| true               | JSON true                                                |
| false              | JSON false                                               |
| null               | JSON null                                                |
| // foo             | inline comment                                           |
| /* foo<br>bar \*/  | multiline comment                                        | 
| // import currency | import statement - see [imported types](#imported-types) |  


### Pre-Defined Types

The following types are pre-defined. some behave differently: 

| Unit  | Example                         | Behaviour                    |
|-------|---------------------------------|------------------------------|
| md    | ```md`hello` ```                | normal - represents markdown |
| json  | ```json`{"foo": [1, 2, 3]}` ``` | value is converted to JSON   |
| yaml  | ```yaml`foo: bar` ```           | value is converted to YAML   |


### Imported Types

Commonly used types can be imported so that they don't need to be defined. For example:

```text
// import currency
$500 £10 50GBP 0.12BTC
```

```json
[
  {"quantity":1,"type":"money","unit":"United States dollar","value":500},
  {"quantity":1,"type":"money","unit":"British pound","value":10},
  {"quantity":50,"type":"money","unit":"British pound"},
  {"quantity":0.12,"type":"money","unit":"Bitcoin"}
]
```


### Injected bb

bb can also be easily extracted and parsed from within comment strings of other language files.

Any comment starting with bb is captured by the parser. For example:

```sql
/*bb
yaml`
  destination: dataset.new_table
  append: false
`
md`# My Amazing Query`
*/

--bb md`Step 1: Select all bars`
SELECT * FROM dataset.table
WHERE foo = 'bar'
```

Then use `--injection-mode` or `-i` when converting to json:

```shell-session
$ bb my-query.sql -i
```

This would return:

```json
[
  {"destination": "dataset.new_table", "append": false},
  {"type": "markdown", "value": "# My Amazing Query"},
  {"type": "markdown", "value": "Step 1: Select all bars"}
]
```

### Examples

The bb: 
```
r = { type: survey response }
4r:no 10r:yes
```
becomes:
 
```json
[{"type":"survey response", "quantity": 4, "value": "no"}, {"type":"survey response", "quantity": 10, "value": "yes"}]
```

Explanation: The value on the left side of the unit is called the **_quantity_** and the value on the right side is the **_value_**. Values can be numbers or quoted strings, or `:` followed by an unquoted string. Quantities can only be numbers. 


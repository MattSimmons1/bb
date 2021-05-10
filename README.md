
# bb

Pictographic programming language. Designed to make data entry super quick and easy!

Define your own types and the syntax for them! For example, the following bb:

    a = { type: apples }
    b = { type: bananas, colour: yellow }
    
    20a 57b
    
Is interpreted as:

```json
[
    { "type": "apple", "quantity": 20 },
    { "type": "banana", "colour": "yellow", "quantity": 57 }
] 
```

Try it out in your browser with the [**bb playground**](https://mattsimmons1.github.io/bb/playground/)!

### Install

If you have go installed you can build the binary yourself and install with:

```bash
go get github.com/MattSimmons1/bb
```

### Usage

```shell-session
$ bb "hello world"  
["hello", "world"]
```

or read from a file:

```shell-session
$ cat my_data.bb.txt
M = { type: message } 
M"hello world" 
$ bb my_data.bb.txt
[{ "type": "message", "value": "hello world" }]
```

### Basic Syntax

| Syntax      | Usage | Result    |
|-------------|-------|-----------|
|number       | 12    | `12`        | 
|string       | foo    | `"foo"`        | 
|safe string  | "foo" or \`foo`   | `"foo"`      | 
|array        | 1 2 "foo"    | `[1, 2, "foo"]` |
|user defined type | ∆ = { }<br>∆ | `{}` | 
|quantity      | ∆ = { }<br>3∆    | `{ "quantity": 3 }` |
|numeric value | ∆ = { }<br>∆5    | `{ "value": 5 }` |
|string value  | ∆ = { }<br>∆\`foo`    | `{ "value": "foo" }` |
|numeric prop  | ∆ = { foo: 100 }<br>∆    |` { "foo": 100 }`   |
|string prop   | ∆ = { foo: bar }<br>∆    | `{ "foo": "bar" }` |
|modifier          | ∆ = { $: foo }<br>∆$1        | `{ "foo": 1 }`          |
|repeated modifier | ∆ = { $: foo }<br>∆$3$\`bar` | `{ "foo": [3, "bar"] }` |
|script prop       | ∆ = { foo: d => 2 * 2 }<br>∆ | `{ "foo": 4 }`          |

### Reserved Characters, Key Words, and Other Syntax

These can't be used as units or modifiers

| Syntax     | Meaning    |
|------------|------------|
| =          | Defines a type |
| .          | Decimal point  |
| -          | Negative sign  |
| { }        | Start and end of a code block or structure |
| true       | JSON true  |
| false      | JSON false |
| null       | JSON null  |
| // foo  | inline comment |
| /* foo<br>bar \*/ | multiline comment | 
| // import currency | import statement - see [imported types](#imported-types)  |  


### Pre-Defined Types

| Unit  | Example | Meaning  |
|-------|---------|----------|
| json  | ```json`{"foo": [1, 2, 3]}```` => `{"foo": [1, 2, 3]}` | The value is parsed as JSON |


### Imported Types

Commonly used types can be optionally imported so that they don't need to be defined. For example:

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
md = { type: markdown }
md`# My Amazing Query`
json`{"destination": "dataset.new_table", "append": false}`
*/

--bb md`Step 1: select all bars`
SELECT * FROM dataset.table
WHERE foo = 'bar'
```

Then use `--injection-mode` or `-i` when converting to json:

```shell-session
$ bb my-query.sql -i
```

### Examples

The bb: 
```
r = { type: survey response }
4r"no" 10r"yes"
```
becomes:
 
```json
[{"type":"survey response", "quantity": 4, "value": "no"}, {"type":"survey response", "quantity": 10, "value": "yes"}]
```
 
Explanation: The value on the left side of the unit is called the **_quantity_** and the value on the right side is the **_value_**. Values can be numbers or quoted strings. Quantities can only be numbers. 


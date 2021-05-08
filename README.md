
# bb

Pictographic programming language. Designed to make data entry super quick and easy!

Define your own types and the syntax for them! For example: The following bb:

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

| bb  | Interpretation  | Explanation  |
|-----|-----------------|--------------| 
| r = { type: survey response }<br>4r"no" 10r"yes" | [{"type":"survey response", "quantity": 4, "value": "no"}, {"type":"survey response", "quantity": 10, "value": "yes"}] | The value on the left side of the unit is called the 'quantity' and the value on the right side is the 'value'. Values can be numbers or quoted strings. Quantities can only be numbers.         |
| y = { type: survey response, "response": "yes" }<br>n = { type: survey response, "response": "no" }<br>4n 10y |  | A different way of representing the same as above, except two types are defined, which allows for easier data entry. |
| y = { type: survey response, "response": "yes" }<br>n = { type: survey response, "response": "no" }<br>yynnyynynyyyyy | [{"type":"survey response", "quantity": 4, "response": "no"}, {"type":"survey response", "quantity": 10, "response": "yes"}] | A different way of representing the same as above, except the order is preserved. |
| ∆ = { food: Pizza, @: price, total: d => d.price * d.quantity }<br>34∆@19.50 | [{"food": "Pizza", "quantity": 34, "price": 19.5, "total": 663 }] | '@' is defined as a modifier. The value following '@' will be the price. | 


### Pre-Defined Types

| Unit  | Example | Meaning  |
|-------|---------|----------|
| json  | ```json`{"foo", [1, 2, 3]}` ``` => `{"foo", [1, 2, 3]}` | The value is parsed as JSON |


### Reserved Characters

These can't be used as units or modifiers

| Character  | Meaning  |
|------------|----------|
| **=**      | Defines a type |
| **.**      | Decimal point  |
| **-**      | Negative sign  |
| **{** **}** | Start and end of a code block or structure |


### Key Words and Other Syntax

| Syntax| Meaning    |
|-------|------------|
| true  | JSON true  |
| false | JSON false |
| null  | JSON null  |
| // foo  | inline comment |
| /* foo<br>bar \*/ | multiline comment | 
| // import currency | import statement - see [imported types](#imported-types)  |  


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

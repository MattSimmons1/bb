
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
| O = { type: Blood Oxygen Level }<br>**O**99 **O**98 **O**85 | [{}]   | Values after the unit become the 'value' property |
| ∆ = { food: Pizza, @: price, total: d => d.price * d.quantity }<br>34∆@19.50 | [{ "food": "Pizza", "quantity": 34, "price": 19.5, "total": 663 }] | '@' is defined as a modifier. The value following '@' will be the price. | 


### Pre-Defined Types

| Unit  | Meaning  |
|-------|----------|
| ·     | null     |


### Reserved Characters

These can't be used as units or modifiers

| Character  | Meaning  |
|------------|----------|
| **=**      | Defines a type |
| **.**      | Decimal point  |
| **-**      | Negative sign  |
| **{** **}** | Start and end of a code block or structure |


### Key Words

| Syntax| Meaning    |
|-------|------------|
| true  | JSON true  |
| false | JSON false |
| null  | JSON null  |


### Other Syntax

| Syntax  | Meaning |
|---------|---------|
| // foo  | inline comment |
| /* foo<br>bar \*/ | multiline comment | 
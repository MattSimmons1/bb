// Type to represent the definition of a UDT. Built-in types are defined here.
package parser

import (
  "strconv"
  "strings"
  "unicode"
)

type udt struct {
  Unit string
  NumericalProps map[string]float64
  StringProps map[string]string
  //JSProps map[string]string  TODO: v0.1.1
  //Specialprops left, right  TODO: v0.1.2 - modifies the left and right arguments
  Modifiers map[rune]Modifier
}

type Modifier struct {  // TODO
  name string
}

func NewUDT(unit string, numericalProps map[string]float64, stringProps map[string]string,
            modifiers map[rune]Modifier) *udt {
  return &udt{ Unit: unit, NumericalProps: numericalProps, StringProps: stringProps, Modifiers: modifiers }
}

// take a definition like "∆ = { "unit": "pizza slices", "+": "topping" }" and add to global map
func NewUDTFromDefinition(definition string) {
  verbose_print("Define new UDT with " + definition)

  // extract just the unit
  i := strings.Index(definition, " ")
  unit := definition[:i]
  definition = definition[i:]

  for {  // remove leading spaces, '=', and '{'
    if definition[0] == ' ' || definition[0] == '=' || definition[0] == '{' {
      definition = definition[1:]
    } else {
      break
    }
  }

  for {  // remove trailing spaces and '}'
    l := len(definition) - 1
    if definition[l] == ' ' || definition[l] == '}' {
      definition = definition[:l]
    } else {
      break
    }
  }

  numericalProps := map[string]float64{}
  stringProps := map[string]string{}
  modifiers := map[rune]Modifier{}

  // TODO: allow commas in strings!
  // split definition into props
  props := strings.Split(definition, ",")

  for _, prop := range props {
    p := strings.SplitN(prop, ":", 2)
    p[0] = strings.TrimSpace(p[0])
    p[1] = strings.TrimSpace(p[1])
    // TODO: remove quotes around prop names

    if len(p[0]) == 1 && isModifierChar(rune(p[0][0])) {
      // TODO: remove quotes around value
      verbose_print("modifier prop: " + p[0] + " = " + p[1])
      modifiers[rune(p[0][0])] = Modifier{ p[1] }
    } else if number, err := strconv.ParseFloat(p[1], 64); err == nil {  // if value is valid number
      verbose_print("numerical prop: " + p[0])
      numericalProps[p[0]] = number
    } else {
      verbose_print("string prop: " + p[0])
      // TODO: remove quotes
      stringProps[p[0]] = p[1]
    }
  }

  verbose_print("unit is " + unit)

  t := NewUDT(unit, numericalProps, stringProps, modifiers)

  UDTs[unit] = t
}

func (t *udt) AddModifier(r rune, name string) {
  t.Modifiers[r] = Modifier{ name }
}

// parse a UDT string - we already know it's valid
func (t *udt) Parse(s string) map[string]interface{} {
  verbose_print("parse " + s + " with unit " + t.Unit)

  pos := strings.Index(s, t.Unit)

  data := make(map[string]interface{})

  quantity := s[0:pos]
  if quantity == "" {
    quantity = "1"
  }
  number, _ := strconv.ParseFloat(quantity, 64)
  data["quantity"] = number

  pos += len(t.Unit)

  if pos != len(s) { // if there is anything remaining

    verbose_print("this is left: " + s[pos:])

    // TODO: parse right value
    if r := rune(s[pos:pos+1][0]); isQuoteChar(r) || isNumeric(r) {
      verbose_print("UDT has value")
      value := ""
      // we already know that value is valid from lexing
      if r == '`' {
        verbose_print("this is left: " + s[pos:])

        valueIdx := pos + 1
        pos = valueIdx
        // TODO: escaped backticks
        for {
          if rune(s[pos]) == '`' {
            break
          } else {
            verbose_print("this is left: " + s[pos:])
            pos++
          }
        }
        value = s[valueIdx:pos]
        pos++
      }
      // TODO: quoted
      // TODO: numerical value

      verbose_print("value is " + value)
      data["value"] = value
    }
  }

  if pos != len(s) {  // if there is anything remaining
    verbose_print("this is left: " + s[pos:])

    allModifiers := ""
    for r := range t.Modifiers {
      allModifiers += string(r)
    }

    verbose_print("looking for modifiers out of: " + allModifiers)

    for {
      if pos == len(s) {
        break
      }
      // next char must be a modifier
      m := t.Modifiers[rune(s[pos:pos+1][0])]
      verbose_print("found modifier: " + m.name)
      pos++
      // find the next modifier
      nextModifierIdx := strings.IndexAny(s[pos:], allModifiers)
      if nextModifierIdx < 0 {
        verbose_print("no more modifiers")
        nextModifierIdx = len(s)  // no more modifiers
      } else {
        nextModifierIdx += pos
      }
      verbose_print(s[pos:])
      verbose_print(s[pos:nextModifierIdx])
      mValue := s[pos:nextModifierIdx]
      if mValue == "" {
        verbose_print("with no value --> 1")
        mValue = "1"
      } else {
        verbose_print("with value: " + mValue)
      }

      if number, err := strconv.ParseFloat(mValue, 64); err == nil { // if value is valid number
        data[m.name] = number
      } else {
        data[m.name] = mValue
      }

      pos = nextModifierIdx
    }

  }


  for k, v := range t.NumericalProps {
    data[k] = v
  }

  for k, v := range t.StringProps {
    data[k] = v
  }

  return data
}


// keeps track of every UDT we have defined
var UDTs = map[string]*udt{
  //"∆": {"∆", map[rune]Modifier{  // e.g.
  //  '+': {"with_breaks"},
  //  '*': {"until_failure"},
  //}},
  //// built-in types
  //"g": {"g", make(map[rune]Modifier)},
  //"kg": {"kg", make(map[rune]Modifier)},

}


var INSTANCES []string  // stores the unit of every UDT we find
var instanceIdx int = 0  // the current index of INSTANCES we're parsing

// identify the type and parse
func ParseUDT(input string) map[string]interface{} {
  unit := INSTANCES[instanceIdx]
  instanceIdx++
  t := UDTs[unit]
  return t.Parse(input)
}

func defineBuiltInTypes() {
  // TODO - find a faster way to define them?
  UDTs["g"] = NewUDT("g", map[string]float64{}, map[string]string{ "unit": "gram" }, map[rune]Modifier{})
}

// return true if a rune could be the start of a udt - slightly faster than checking the whole thing?
func couldBeUDT(r rune) bool {
  if unicode.IsDigit(r) {
    return true
  }

  i := 0
  for k := range UDTs {
    verbose_print(string([]rune(k)[0]))
    if r == []rune(k)[0] {  // does the first rune in the UDT's unit match
      verbose_print("could " + string(r) + " be a udt? yes!")
      return true
    }
    i++
  }
  verbose_print("could " + string(r) + " be a udt? no.")
  return false
  //_, ok := UDTs[string(r)]
  //if !ok {
  //
  //}
  //return ok
}
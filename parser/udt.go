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
  Modifiers map[rune]Modifier
}

type Modifier struct {
  name string
}

func NewUDT(unit string, numericalProps map[string]float64, stringProps map[string]string,
            modifiers map[rune]Modifier) *udt {
  return &udt{ Unit: unit, NumericalProps: numericalProps, StringProps: stringProps, Modifiers: modifiers }
}

// take a definition like "∆ = { "unit": "pizza slices", "+": "topping" }" and add to global map
func NewUDTFromDefinition(definition string) {
  log("Define new UDT with " + definition)

  // extract just the unit
  i := strings.Index(definition, " ")  // TODO: allow no spaces
  if i == -1 {
    i = strings.Index(definition, "=")
  }
  unit := definition[:i]
  log("unit is " + unit)
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
    if l == -1 {
      break  // definition is nil
    } else if definition[l] == ' ' || definition[l] == '}' {
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
    if len(p) < 2 {
      log("INVALID PROPS")
      continue
    }
    p[0] = strings.TrimSpace(p[0])
    p[1] = strings.TrimSpace(p[1])
    // TODO: remove quotes around prop names

    if len(p[0]) == 1 && isModifierChar(rune(p[0][0])) {
      // TODO: remove quotes around value
      log("modifier prop: " + p[0] + " = " + p[1])
      modifiers[rune(p[0][0])] = Modifier{ p[1] }
    } else if number, err := strconv.ParseFloat(p[1], 64); err == nil {  // if value is valid number
      log("numerical prop: " + p[0])
      numericalProps[p[0]] = number
    } else {
      log("string prop: " + p[0])
      // TODO: remove quotes
      stringProps[p[0]] = p[1]
    }
  }

  t := NewUDT(unit, numericalProps, stringProps, modifiers)

  UDTs[unit] = t
}

func (t *udt) AddModifier(r rune, name string) {
  t.Modifiers[r] = Modifier{ name }
}

// parse a UDT string - we already know it's valid
func (t *udt) Parse(s string) map[string]interface{} {
  log("parse " + s + " with unit " + t.Unit)

  pos := strings.Index(s, t.Unit)

  data := make(map[string]interface{})

  quantity := s[0:pos]
  if quantity == "" {
    quantity = "1"
  }
  number, _ := strconv.ParseFloat(quantity, 64)
  data["quantity"] = number

  pos += len(t.Unit)

  length := len(s)

  if pos != length { // if there is anything remaining

    log("this is left: " + s[pos:])

    if r := rune(s[pos:pos+1][0]); isQuoteChar(r) {
      log("UDT has value")
      // we already know that value is valid from lexing
      if isQuoteChar(r) {
        log("this is left: " + s[pos:])

        valueIdx := pos + 1
        pos = valueIdx
        for {
          if rune(s[pos]) == r {
            break
          } else if rune(s[pos]) == '\\' && rune(s[pos+1]) == r  {
              pos+=2  // escaped quotes
          } else {
            log("this is left: " + s[pos:])
            pos++
          }
        }
        data["value"] = s[valueIdx:pos]
        pos++
      }

      log("value is " + data["value"].(string))

    } else if isNumeric(r) {
      valueIdx := pos
      pos = valueIdx
      for {
        if pos == length || !isNumeric(rune(s[pos])) {
          break
        } else {
          log("this is left: " + s[pos:])
          pos++
        }
      }
      number, err := strconv.ParseFloat(s[valueIdx:pos], 64)
      if err != nil {
        panic(err)
      }
      data["value"] = number
    }
  }

  if pos != length {  // if there is anything remaining
    log("this is left: " + s[pos:])

    allModifiers := ""
    for r := range t.Modifiers {
      allModifiers += string(r)
    }

    log("looking for modifiers out of: " + allModifiers)

    for {
      if pos == len(s) {
        break
      }
      // next char must be a modifier
      m := t.Modifiers[rune(s[pos:pos+1][0])]
      log("found modifier: " + m.name)
      pos++
      // find the next modifier
      nextModifierIdx := strings.IndexAny(s[pos:], allModifiers)
      if nextModifierIdx < 0 {
        log("no more modifiers")
        nextModifierIdx = len(s)  // no more modifiers
      } else {
        nextModifierIdx += pos
      }
      log(s[pos:])
      log(s[pos:nextModifierIdx])
      mValue := s[pos:nextModifierIdx]
      if mValue == "" {
        log("with no value --> 1")
        mValue = "1"
      } else {
        log("with value: " + mValue)
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

var UDTs = map[string]*udt{}  // stores user defined types
var PDTs = map[string]*udt{}  // stores pre-defined types

var INSTANCES []string  // stores the unit of every UDT we find
var instanceIdx int = 0  // the current index of INSTANCES we're parsing

// identify the type and parse
func ParseUDT(input string) map[string]interface{} {
  unit := INSTANCES[instanceIdx]
  instanceIdx++
  t := UDTs[unit]
  if t == nil {
    t = PDTs[unit]
  }
  return t.Parse(input)
}

func defineBuiltInTypes() {
  //SITypes := []string{ - TODO: make optional
  //  "g,gram,weight",
  //  "kg,kilogram,weight",
  //  "s,second,time",
  //  "min,minute,time",
  //  "h,hour,time",
  //  "d,day,time",
  //  "m,metre,length",
  //  "km,kilometre,length",
  //  "au,astronomical unit,length",
  //  "l,litre,volume",
  //  "K,kelvin,temperature",
  //  "A,ampere,electric current",
  //  "mol,mole,amount of substance",
  //  "cd,candela,luminous intensity",
  //  "rad,radian,plane angle",
  //  "sr,steradian,solid angle",
  //  "Hz,hertz,frequency",
  //  "N,newton,force",
  //  "Pa,pascal,pressure",
  //  "J,joule,energy",
  //  "eV,electron volt,energy",
  //  "W,watt,power",
  //  "C,coulomb,electric charge",
  //  "V,volt,voltage",
  //  "F,farad,capacitance",
  //  "Ω,ohm,resistance",
  //  "S,siemens,electrical conductance",
  //  "Wb,weber,magnetic flux",
  //  "T,tesla,magnetic flux density",
  //  "H,henry,inductance",
  //  "°C,Celsius,temperature",
  //  "lm,lumen,luminous flux",
  //  "lx,lux,illuminance",
  //  "Bq,becquerel,radioactivity",
  //  "Gy,gray,absorbed dose",
  //  "Sv,sievert,equivalent dose",
  //  "kat,katal,catalytic activity",
  //}
  //
  //for _, t := range SITypes {
  //  def := strings.SplitN(t, ",", 3)
  //  PDTs[def[0]] = NewUDT(def[0], map[string]float64{}, map[string]string{ "unit": def[1], "type": def[2] }, map[rune]Modifier{})
  //}

  //Quantity	Name	Symbol	Value in SI units
  //plane and
  //phase angle	degree	°	1° = (π/180) rad
  //minute	′	1′ = (1/60)° = (π/10800) rad
  //second	″	1″ = (1/60)′ = (π/648000) rad
  //area	hectare	ha	1 ha = 1 hm2 = 104 m2
  //mass	tonne (metric ton)	t	1 t = 1 000 kg
  //dalton	Da	1 Da = 1.660539040(20)×10−27 kg
  //bel	B
  //decibel	dB

  // currency - TODO: make optional
  //currencyTypes := []string{
  //  "$,United States dollar",
  //  "£,Great British pound sterling",
  //  "¥,Japanese yen",
  //  "円,Japanese yen",
  //  "元,Chinese renminbi yuan",
  //}
  //for _, t := range currencyTypes {
  //  def := strings.SplitN(t, ",", 2)
  //  PDTs[def[0]] = NewUDT(def[0], map[string]float64{}, map[string]string{ "unit": def[1], "type": "money" }, map[rune]Modifier{})
  //}
}

// return true if a rune could be the start of a udt - slightly faster than checking the whole thing
func couldBeUDT(r rune) bool {
  if unicode.IsDigit(r) {
    return true
  }

  i := 0
  for k := range UDTs {  // check UDTs first
    log(string([]rune(k)[0]))
    if r == []rune(k)[0] {  // does the first rune in the UDT's unit match
      log("could " + string(r) + " be a udt? yes!")
      return true
    }
    i++
  }
  for k := range PDTs {  // now check PDTs TODO: pre-make this array?
    log(string([]rune(k)[0]))
    if r == []rune(k)[0] {  // does the first rune in the UDT's unit match
      log("could " + string(r) + " be a udt? actually it's a pdt!")
      return true
    }
  }
  log("could " + string(r) + " be a udt? no.")
  return false
}
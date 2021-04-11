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
  ScriptProps map[string]string
  HiddenProps []string
}

func NewUDT(unit string, numericalProps map[string]float64, stringProps map[string]string,
            scriptProps map[string]string) *udt {
  return &udt{ Unit: unit, NumericalProps: numericalProps, StringProps: stringProps, ScriptProps: scriptProps }
}

// take a definition like "∆ = { "unit": "pizza slices", "+": "topping" }" and add to global map
func NewUDTFromDefinition(definition string) {
  log("Define new UDT with " + definition)

  // extract just the unit
  i := strings.Index(definition, "=")
  unit := strings.TrimSpace(definition[:i])
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
  scriptProps := map[string]string{}

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

    //if len(p[0]) == 1 && isModifierChar(rune(p[0][0])) {
      //log("modifier prop: " + p[0] + " = " + p[1])
      //modifiers[rune(p[0][0])] = Modifier{ p[1] }
    //} else
    if number, err := strconv.ParseFloat(p[1], 64); err == nil {  // if value is valid number
      log("numerical prop: " + p[0])
      numericalProps[p[0]] = number
    } else if strings.Contains(p[1], "=>") {  // if value is an arrow function - TODO: check for single left hand argument and don't match strings that contain => but aren't functions
      log("script prop: " + p[0] + ", with value: " + p[1])
      functionStart := strings.Index(p[1], "=>")
      // we must re-write as a normal function because we can only run ES5 syntax 
      function := "function f(" + p[1][:functionStart] + "){ return " + p[1][functionStart+2:] + "};"
      log(function)
      scriptProps[p[0]] = function
    } else {
      log("string prop: " + p[0])
      // TODO: remove quotes
      stringProps[p[0]] = removeQuotes(p[1])
    }
  }

  t := NewUDT(unit, numericalProps, stringProps, scriptProps)

  UDTs[unit] = t
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
  number, err := strconv.ParseFloat(quantity, 64)
  if err != nil {
    data["quantity"] = quantity  // invalid quantities are kept as string, e.g. 1.0.0
  } else {
    data["quantity"] = number
  }

  pos += len(t.Unit)

  length := len(s)

  if pos != length { // if there is anything remaining

    if r := rune(s[pos]); isQuoteChar(r) {
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
          pos++
        }
      }
      number, err := strconv.ParseFloat(s[valueIdx:pos], 64)
      if err != nil {
        data["value"] = s[valueIdx:pos]  // invalid values are kept as string, e.g. 1.0.0
      } else {
        data["value"] = number
      }
    }
  }

  for modifier, rawValue := range *MODIFIER_INSTANCES[instanceIdx] {
    t.addModifierToData(data, modifier, rawValue)
  }

  //if pos != length {  // if there is anything remaining
  //  log("looking for modifiers")
  //  log("this is left: " + s[pos:])
  //
  //  // bar"wooie"b"wool"
  //  // scan ahead to the next non modifier character
  //  modifierStart := pos
  //  for {
  //    if pos == length { // when we've reached the end - check for modifiers with no value - then stop
  //      if pos != modifierStart {
  //        log("modifier with no value")
  //        m := s[modifierStart:pos]
  //        t.addModifierToData(data, m, "1")
  //      }
  //      break
  //    }
  //
  //    r := rune(s[pos])
  //
  //    if isSpace(r) || isNumeric(r) || r == '"' || r == '`' || r == '.' || r == '-' {
  //
  //      // TODO check for - or . modifiers
  //      m := s[modifierStart:pos]
  //      backtrackCharacters := 0
  //
  //      backtrackLoop: for {
  //        if len(m) == backtrackCharacters {
  //          break backtrackLoop
  //        }
  //
  //        m2 := m[:len(m)-backtrackCharacters]
  //
  //        log("the modifier unit could be " + m2)
  //
  //        if m == "" {
  //          panic("Parsing Error") // TODO
  //        }
  //
  //        for modifier := range t.StringProps {
  //          if m2 == modifier {
  //            log("modifier is: \033[92m" + m2 + "\033[0m")
  //            //pos -= backtrackCharacters
  //
  //            // loop through the value
  //            valueStart := pos
  //
  //            quoted := false
  //            quoteChar := rune(s[pos])
  //            if quoteChar == '"' || quoteChar == '`' {
  //              log("value is quoted")
  //              quoted = true
  //              pos++
  //            }
  //
  //          Loop:
  //            for {
  //              if pos == length {
  //                break Loop
  //              }
  //              switch r := rune(s[pos]); {
  //              case isNumeric(r):
  //                pos++ // absorb
  //                log(string(r))
  //
  //              case quoted && r == '\\':
  //                if rune(s[pos+1]) == quoteChar {
  //                  pos += 2 // absorb escaped quote
  //                  log("found escaped quote")
  //                } else {
  //                  pos++  // backslash is absorbed
  //                }
  //              case quoted && r != quoteChar && r != '\n':
  //                pos++ // absorb
  //                log(string(r))
  //
  //              default:
  //                break Loop
  //              }
  //            }
  //            mValue := ""
  //            if quoted {
  //              mValue = s[valueStart+1 : pos]
  //              pos++ // absorb the quote char
  //            } else {
  //              mValue = s[valueStart:pos]
  //            }
  //            log("value is " + s[valueStart:pos])
  //            t.addModifierToData(data, modifier, mValue)
  //
  //          }
  //        }
  //        backtrackCharacters++
  //
  //      }
  //      modifierStart = pos // onto the next modifier
  //      // if no modifier was found - it's an error
  //      //panic("no modifier matched - THIS SHOULD NOT HAPPEN!!!")
  //
  //    } else {
  //      log(string(r))
  //      pos += 1
  //    }
  //
  //  }
  //
  //}

  for k, v := range t.NumericalProps {
    data[k] = v
  }

  LoopStringProps: for k, v := range t.StringProps {
    for _, prop := range t.HiddenProps {
      if prop == k {
        continue LoopStringProps  // skip props that should be hidden
      }
    }
    if isModifierChar(rune(k[0])) {
      continue LoopStringProps  // skip props that start with standard modifier chars TODO: or give them null value?
    }
    data[k] = v
  }

  for k, v := range t.ScriptProps {
    data[k] = RunScript(v, data)
  }

  return data
}

func (t *udt) addModifierToData(data map[string]interface{}, modifier string, value string) {
  t.HiddenProps = append(t.HiddenProps, modifier)
  modifierName := t.StringProps[modifier]
  appendValue := false

  if value == "" {
    value = "1"
  }

  // remove quotes
  value = removeQuotes(value)

  if data[modifierName] != nil {
    if data[modifierName] == "" {  // TODO: why does this happen?
      // don't append
    } else {
      appendValue = true
      if _, ok := data[modifierName].([]interface{}); !ok {  // if the value is not already a slice
        data[modifierName] = []interface{}{ data[modifierName] }  // convert to slice
      }
    }
  }
  if number, err := strconv.ParseFloat(value, 64); err == nil { // if value is valid number
    if appendValue {
      data[modifierName] = append(data[modifierName].([]interface{}), number)
    } else {
      data[modifierName] = number
    }
  } else {
    if appendValue {
      data[modifierName] = append(data[modifierName].([]interface{}), value)
    } else {
      data[modifierName] = value
    }
  }
}

var UDTs = map[string]*udt{}  // stores user defined types
var PDTs = map[string]*udt{}  // stores pre-defined types

var INSTANCES []string  // stores the unit of every UDT we find
var MODIFIER_INSTANCES []*map[string]string  // stores every modifier and raw value we find
var instanceIdx int = 0  // the current index of INSTANCES we're parsing

// identify the type and parse
func ParseUDT(input string) map[string]interface{} {
  unit := INSTANCES[instanceIdx]

  f := func () {
    instanceIdx++
  }
  defer f()

  t := UDTs[unit]
  if t == nil {
    t = PDTs[unit]
  }
  return t.Parse(input)
}

func defineBuiltInTypes(collectionName string) {

  if collectionName == "si" {
    log("Importing SI Units")

    SITypes := []string{
      "g,gram,weight",
      "kg,kilogram,weight",
      "s,second,time",
      "min,minute,time",
      "h,hour,time",
      "d,day,time",
      "m,metre,length",
      "km,kilometre,length",
      "au,astronomical unit,length",
      "l,litre,volume",
      "K,kelvin,temperature",
      "A,ampere,electric current",
      "mol,mole,amount of substance",
      "cd,candela,luminous intensity",
      "rad,radian,plane angle",
      "sr,steradian,solid angle",
      "Hz,hertz,frequency",
      "N,newton,force",
      "Pa,pascal,pressure",
      "J,joule,energy",
      "eV,electron volt,energy",
      "W,watt,power",
      "C,coulomb,electric charge",
      "V,volt,voltage",
      "F,farad,capacitance",
      "Ω,ohm,resistance",
      "S,siemens,electrical conductance",
      "Wb,weber,magnetic flux",
      "T,tesla,magnetic flux density",
      "H,henry,inductance",
      "°C,Celsius,temperature",
      "lm,lumen,luminous flux",
      "lx,lux,illuminance",
      "Bq,becquerel,radioactivity",
      "Gy,gray,absorbed dose",
      "Sv,sievert,equivalent dose",
      "kat,katal,catalytic activity",
    }

    for _, t := range SITypes {
     def := strings.SplitN(t, ",", 3)
     PDTs[def[0]] = NewUDT(def[0], map[string]float64{}, map[string]string{ "unit": def[1], "type": def[2] },
                           map[string]string{})
    }
  }

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

  if collectionName == "currency" || collectionName == "money" {
    currencyTypes := []string{
      "$,USD,United States dollar",
      "£,GBP,British pound",
      "€,EUR,Euro",
      "¥,JPY,Japanese yen",
      "円,Japanese yen",
      "元,Chinese renminbi yuan",
      "₹,Indian rupee",
      "₽,RUB,Russian ruble",
      "฿,Thai baht",

      // crypto
      "₿,BTC,Bitcoin",
      "ETH,Ether",
      "Ł,LTE,Litecoin",
      "₳,ADA,Ada",
    }
    for _, t := range currencyTypes {
     def := strings.Split(t, ",")
     for _, unit := range def[:len(def)-1] {
       PDTs[unit] = NewUDT(unit, map[string]float64{}, map[string]string{ "unit": def[len(def)-1], "type": "money" },
                           map[string]string{})
     }
    }
  }

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
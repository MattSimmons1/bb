package parser

import (
	"encoding/json"
	"testing"
)

type testCase struct {
	name   string
	raw    string
	parsed string
}

var testCases = []testCase{
	{"simple udt", "a = { b: c }\n1a2 a3b a4b5", `[{"b":"c","quantity":1,"value":2},{"c":true,"value":3},{"c":5,"value":4}]`},
	{"simple no spaces", "a={b:c}\n1a2", `[{"b":"c","quantity":1,"value":2}]`},
	{"comment in definition", "a={//foo\nb:c}\na2", `[{"b":"c","value":2}]`},
	{"quote modifier", "a={\":c}\na2\" a\"foo\"\" a\"", `[{"c":true,"value":2},{"c":true,"value":"foo"},{"c":true}]`},
	{"udt-like strings", "x 1234 z 12x 213 34x", `["x",1234,"z","12x",213,"34x"]`},
	{"negative numbers", "-1 -0.1 -.0 .2 -12x", `[-1,-0.1,-0,0.2,"-12x"]`},
	{"unquoted values", `a = {} a:helloa:`, `[{"value":"helloa:"}]`},
	{"unquoted values disabled", `: = { t: u } a = {} a:hello: a:a`, `[{},{"t":"u"},"hello:",{},{"t":"u"},{}]`},
	{"repeated modifier", `∆ = {+:f} ∆+3+"b"`, `[{"f":[3,"b"]}]`},
	{"repeated modifier bool", `∆ = {+:f} ∆+++`, `[{"f":[true,true,true]}]`},
	//{"decimals_disabled", `. = {} 1.2.`, ``},
	// TODO: how to allow for other types but not this one? what should happen?
	//{"dot modifier", `a = { .: b } 1.1a."a" a.1.1`, `[{"b":"a","quantity":1.1}]`},
	// TODO: not sure if this is what we want to happen - disable negative numbers when '-' is a modifier
	{"dash modifier", `a = { -: b } -1a-2-3`, `[{"b":3,"quantity":-1,"value":-2}]`},
	{"si units", "// import si\n50g 234T 23Bq 77l", `[{"quantity":50,"type":"weight","unit":"gram"},{"quantity":234,"type":"magnetic flux density","unit":"tesla"},{"quantity":23,"type":"radioactivity","unit":"becquerel"},{"quantity":77,"type":"volume","unit":"litre"}]`},
	{"currency", "// import currency\n$500 £10 50GBP 0.12BTC", `[{"type":"money","unit":"United States dollar","value":500},{"type":"money","unit":"British pound","value":10},{"quantity":50,"type":"money","unit":"British pound"},{"quantity":0.12,"type":"money","unit":"Bitcoin"}]`},
	{"script props", `∆={g:g,f:d =>d.value+d.g} ∆1g3 ∆"goo"g"foo" ∆"ya"g0g1g2 ∆`, `[{"f":4,"g":3,"value":1},{"f":"goofoo","g":"foo","value":"goo"},{"f":"ya0,1,2","g":[0,1,2],"value":"ya"},{"f":null}]`},

	//{"modifiers", "a = { t: a, *: b, !: c }\na* a*2 2a** a*! a!!`yes`!`no`", ``},  // TODO: not working properly
	//{"dates", "2016-01-01", `[{"type": "date", value: "2016-01-01"}]`},  // TODO: future feature
}

func Test_Parse_simple(t *testing.T) {

	verbose = true

	for i := range testCases {

		data := Parse(testCases[i].raw)
		result, err := json.Marshal(data)

		if err != nil {
			t.Fatalf(`Couldn't parse output as json: %s`, err)
		}

		if string(result) != testCases[i].parsed {
			t.Fatalf(`Failed test case '%s': 
%s 
--> Not the expected output: %s vs %s`, testCases[i].name, testCases[i].raw, result, testCases[i].parsed)
		}

	}
}

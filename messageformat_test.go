package icu

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestTranslate(t *testing.T) {
	testCases := []struct {
		name       string
		message    MessageFormat
		parameters []Parameter
		translated string
	}{
		{"text", "Hello!", nil, "Hello!"},
		{"text:with-parameter", "Hello!", []Parameter{P("foo", "Bar")}, "Hello!"},
		{"placeholder:one-value", "{name}", []Parameter{P("name", "Foo")}, "Foo"},
		{"placeholder:one-value:empty", "{name}", nil, ""},
		{"placeholder:one-value-text", "Hello {name}!", []Parameter{P("name", "Bob")}, "Hello Bob!"},
		{"placeholder:one-value-text:empty", "Hello {name}!", nil, "Hello !"},
		{"placeholder:two-value", "{given-name} {family-name}", []Parameter{P("given-name", "Mario"), P("family-name", "Demuth")}, "Mario Demuth"},
		{"placeholder:two-value:reorder", "{family-name}, {given-name}", []Parameter{P("given-name", "Mario"), P("family-name", "Demuth")}, "Demuth, Mario"},
		{"placeholder:two-value:multi-use", "{given-name} {given-name} {given-name} {family-name}", []Parameter{P("given-name", "Mario"), P("family-name", "Demuth")}, "Mario Mario Mario Demuth"},
		{"number", "{value, number, %.0f}", []Parameter{P("value", 5.4)}, "5"},
		{"number:round", "{value, number, %.0f}", []Parameter{P("value", 5.5)}, "6"},
		{"select", "{gender, select, male {♂} female {♀}}", []Parameter{P("gender", "male")}, "♂"},
		{"select", "{gender, select, male {♂} female {♀}}", []Parameter{P("gender", "female")}, "♀"},
		{"select", "{gender, select, - {} male {♂} other {♀}}", []Parameter{P("gender", "female")}, "♀"},
		{"select", "{gender, select, - {} male {♂} other {♀}}", []Parameter{P("gender", "-")}, ""},
		{"select:bool", "{has-train-number, select, true {Train-Number: {train-number}} other {Train-Number: not available}}", []Parameter{P("has-train-number", true), P("train-number", "B8")}, "Train-Number: B8"},
		{"select:bool", "{has-train-number, select, true {Train-Number: {train-number}} other {Train-Number: not available}}", []Parameter{P("has-train-number", false)}, "Train-Number: not available"},
		// {"convert:celsius-to-fahrenheit", "{value, celsius-to-fahrenheit, %.0f}", []Parameter{P("value", 0)}, "32"},
		// {"convert:celsius-to-kelvin", "{value, celsius-to-kelvin, %.2f}", []Parameter{P("value", 0)}, "273.15"},
		// {"convert:kmh-to-mph", "{value, kmh-to-mph, %.3f}", []Parameter{P("value", 1)}, "0.621"},
		// {"convert:kmh-to-mph", "{value, kmh-to-mph, %.3f}", []Parameter{P("value", 1.0)}, "0.621"},
		// {"convert:kn-to-kips", "{value, kilonewton-to-kip, %.4f}", []Parameter{P("value", 1)}, "0.2248"},
		// {"convert:kn-to-kips", "{value, kilonewton-to-kip, %.4f}", []Parameter{P("value", 1.0)}, "0.2248"},
		{"complex", "Hello {name}, you are a {gender, select, male {♂} female {♀}}.", []Parameter{P("gender", "male"), P("name", "Mario")}, "Hello Mario, you are a ♂."},
		{"complex:male", "{gender, select, male {{name} is ♂ and he is {age} years old.} female {{name} is ♀ and she is {age} years old.}}", []Parameter{P("gender", "male"), P("name", "Mario"), P("age", 25)}, "Mario is ♂ and he is 25 years old."},
		{"complex:female", "{gender, select, male {{name} is ♂ and he is {age} years old.} female {{name} is ♀ and she is {age} years old.}}", []Parameter{P("gender", "female"), P("name", "Alice"), P("age", 17)}, "Alice is ♀ and she is 17 years old."},

		{"quoted", "Use '{foo}' as a variable.", nil, "Use {foo} as a variable."},
		{"quoted", "Use # as a text.", nil, "Use # as a text."},
		{"quoted", "Value # as a text.", []Parameter{P("value", 1)}, "Value 1 as a text."},
		{"quoted", "Use # as a text.", []Parameter{P("$value", "foo")}, "Use # as a text."},

		//{"quoted:quote", "This '{isn''t}' obvious", nil, "This {isn't} obvious"},

		{"plural:zero", pluralCardinal, []Parameter{P("count", 0)}, "   no \n alarms \t where issued "},
		{"plural:one", pluralCardinal, []Parameter{P("count", 1)}, "  one alarm was issued  "},
		{"plural:other", pluralCardinal, []Parameter{P("count", 3)}, " 3 alarms where issued   "},

		{"plural:offset:0", pluralWithOffset, []Parameter{P("count", 0)}, "no alarms where issued"},
		{"plural:offset:1", pluralWithOffset, []Parameter{P("count", 1)}, "one alarm was issued"},
		{"plural:offset:one", pluralWithOffset, []Parameter{P("count", 2)}, "another alarm was issued"},
		{"plural:offset:other", pluralWithOffset, []Parameter{P("count", 3)}, "2 alarms where issued"},

		{"selectOrdinal:one", pluralOrdinal, []Parameter{P("num", 1)}, "1st"},
		{"selectOrdinal:two", pluralOrdinal, []Parameter{P("num", 2)}, "2nd"},
		{"selectOrdinal:few", pluralOrdinal, []Parameter{P("num", 33)}, "33rd"},
		{"selectOrdinal:oth", pluralOrdinal, []Parameter{P("num", 14)}, "14th"},

		{"party", party, []Parameter{P("gender_of_host", "female"), P("num_guests", 0), P("host", "Alice"), P("guest", "Bob")}, "Alice does not give a party."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Translate("en", tc.message, tc.parameters...)
			if err != nil {
				t.Errorf("parse: %s", err)
			}

			if tc.translated != got {
				t.Errorf("expected: '%s', got: '%s'", tc.translated, got)
			}
		})
	}

}

const pluralCardinal = "{count, plural," +
	"=0 {   no \n alarms \t where issued }" +
	"=1 {  one alarm was issued  }" +
	"other { {count} alarms where issued   }" +
	"}"

const pluralOrdinal = "{num, selectordinal," +
	" one {#st}" +
	" two {#nd}" +
	" few {#rd}" +
	" other {#th}" +
	"}"

const pluralWithOffset = "{count, plural, offset:1" +
	" =0 {no alarms where issued}" +
	" =1 {one alarm was issued}" +
	" one {another alarm was issued}" +
	" other{# alarms where issued}" +
	"}"

const party = "{gender_of_host, select," +
	" female {" +
	"{num_guests, plural, offset:1" +
	" =0 {{host} does not give a party.}" +
	" =1 {{host} invites {guest} to her party.}" +
	" =2 {{host} invites {guest} and one other person to her party.}" +
	" other {{host} invites {guest} and # other people to her party.}" +
	"}" +
	"}" +
	" male {" +
	"{num_guests, plural, offset:1" +
	" =0 {{host} does not give a party.}" +
	" =1 {{host} invites {guest} to his party.}" +
	" =2 {{host} invites {guest} and one other person to his party.}" +
	" other {{host} invites {guest} and # other people to his party.}" +
	"}" +
	"}" +
	"other {" +
	"{num_guests, plural, offset:1" +
	" =0 {{host} does not give a party.}" +
	" =1 {{host} invites {guest} to their party.}" +
	" =2 {{host} invites {guest} and one other person to their party.}" +
	" other {{host} invites {guest} and # other people to their party.}" +
	"}" +
	"}" +
	"}"

func TestParametersFrom(t *testing.T) {
	testCases := []struct {
		from       interface{}
		parameters Parameters
	}{
		{
			from:       nil,
			parameters: nil,
		},
		{
			from: map[string]string{
				"foo": "bar",
			},
			parameters: Parameters{
				P("foo", "bar"),
			},
		},
		{
			from: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
			parameters: Parameters{
				P("bar", "baz"),
				P("foo", "bar"),
			},
		},
		{
			from: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			parameters: Parameters{
				P("bar", "baz"),
				P("foo", "bar"),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			sort.Sort(tc.parameters)
			got := ParametersFrom(tc.from)
			sort.Sort(got)
			if !reflect.DeepEqual(tc.parameters, got) {
				t.Errorf("want: %v, got: %v", tc.parameters, got)
			}
		})
	}
}

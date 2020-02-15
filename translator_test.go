package icu

import "testing"

func TestTranslator(t *testing.T) {
	trans := NewHierachicalTranslator()
	trans.Translations["foo:bar"] = "Hello {foo}!"
	trans.Translations["zap"] = "{fizzle}"

	want := "Hello FOO!"
	got := trans.Translate("foo:bar", P("foo", "FOO"))
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}

	want = "BAR"
	got = trans.Translate("zap", P("fizzle", "BAR"))
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}

	want = "fizzle"
	got = trans.Translate("fizzle")
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

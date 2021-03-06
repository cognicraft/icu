package icu

type Translator interface {
	Translate(key string, ps ...Parameter) string
}

type TranslatorFunc func(key string, ps ...Parameter) string

func (f TranslatorFunc) Translate(key string, ps ...Parameter) string {
	return f(key, ps...)
}

func NewHierachicalTranslator() *HierachicalTranslator {
	return &HierachicalTranslator{
		Tag:          TagEn,
		Translations: map[string]MessageFormat{},
	}
}

type HierachicalTranslator struct {
	Base         Translator
	Tag          Tag
	Translations map[string]MessageFormat
}

func (t *HierachicalTranslator) IsRoot() bool {
	return t == nil || t.Base == nil
}

func (t *HierachicalTranslator) Translate(key string, ps ...Parameter) string {
	if t == nil {
		return key
	}
	if mf, ok := t.Translations[key]; ok {
		if v, err := mf.Translate(t.Tag, ps...); err == nil {
			return v
		}
	}
	if !t.IsRoot() {
		return t.Base.Translate(key, ps...)
	}
	return key
}

var nilTranslator Translator = TranslatorFunc(nilTranslatorFunc)

func nilTranslatorFunc(key string, ps ...Parameter) string {
	return key
}

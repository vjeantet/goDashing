package gerb

import (
	"crypto/sha1"
	"fmt"
	"github.com/karlseguin/gerb/core"
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/ccache.v1"
	"io"
	"io/ioutil"
	"time"
)

var Cache = ccache.New(ccache.Configure().MaxSize(5000))

// a chain of templates to render from inner-most to outer-most
type TemplateChain []core.Executable

func (t TemplateChain) Render(writer io.Writer, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	defaultContent := core.BytePool.Checkout()
	context := &core.Context{
		Writer:   defaultContent,
		Data:     data,
		Counters: make(map[string]int),
		Contents: map[string]*bytepool.Bytes{"$": defaultContent},
	}
	defer cleanup(context)
	lastIndex := len(t) - 1
	for i := 0; i < lastIndex; i++ {
		t[i].Execute(context)
	}
	context.Writer = writer
	t[lastIndex].Execute(context)
}

// Parse the bytes into a gerb template
func Parse(cache bool, data ...[]byte) (TemplateChain, error) {
	templates := make(TemplateChain, len(data))
	for index, d := range data {
		template, err := parseOne(cache, d)
		if err != nil {
			return nil, err
		}
		templates[index] = template
	}
	return templates, nil
}

// Parse the string into a erb template
func ParseString(cache bool, data ...string) (TemplateChain, error) {
	all := make([][]byte, len(data))
	for index, d := range data {
		all[index] = []byte(d)
	}
	return Parse(cache, all...)
}

// Turn the contents of the specified file into a gerb template
func ParseFile(cache bool, paths ...string) (TemplateChain, error) {
	all := make([][]byte, len(paths))
	for index, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		all[index] = data
	}
	return Parse(cache, all...)
}

func parseOne(cache bool, data []byte) (*Template, error) {
	if cache == false {
		return newTemplate(data)
	}
	hasher := sha1.New()
	hasher.Write(data)
	key := fmt.Sprintf("%x", hasher.Sum(nil))

	t := Cache.Get(key)
	if t != nil {
		return t.Value().(*Template), nil
	}

	template, err := newTemplate(data)
	if err != nil {
		return nil, err
	}
	Cache.Set(key, template, time.Hour)
	return template, nil
}

func cleanup(context *core.Context) {
	for _, writer := range context.Contents {
		writer.Close()
	}
}

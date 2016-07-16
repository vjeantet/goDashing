package gerb

import (
	"errors"
	"github.com/karlseguin/gerb/core"
)

func newTemplate(data []byte) (*Template, error) {
	template := &Template{new(core.NormalContainer)}
	stack := []core.Code{template}
	var container core.Code = template
	parser := core.NewParser(data)
	trim := false
	for {
		if literal := parser.ReadLiteral(trim); literal != nil {
			container.AddExecutable(literal)
		}
		trim = false
		tagType := parser.ReadTagType()
		if tagType == core.NoTag {
			return template, nil
		}

		isUnsafe := tagType == core.UnsafeTag
		if tagType == core.OutputTag || isUnsafe {
			output, err := createOutputTag(parser, isUnsafe)
			if err != nil {
				return nil, err
			}
			if output != nil {
				container.AddExecutable(output)
			}
		} else if tagType == core.CommentTag {
			t, err := parser.ReadComment()
			if err != nil {
				return nil, err
			}
			trim = t
		} else if tagType == core.CodeTag {
			codes, err := createCodeTag(parser)
			if err != nil {
				return nil, err
			}
			if codes != nil {
				trim = codes.Trim
				for _, code := range codes.List {
					if code == endScope {
						l := len(stack) - 1
						stack = stack[0:l]
						container = stack[l-1]
					} else {
						if code.IsSibling() {
							if err := stack[len(stack)-1].AddCode(code); err != nil {
								return nil, err
							}
						} else {
							container.AddExecutable(code)
						}
						if code.IsContentContainer() {
							container = code
						}
						if code.IsCodeContainer() {
							stack = append(stack, container)
						}
					}
				}
			}
		}
	}
	return template, nil
}

type Template struct {
	*core.NormalContainer
}

func (t *Template) IsCodeContainer() bool {
	return true
}

func (t *Template) IsContentContainer() bool {
	return true
}

func (t *Template) IsSibling() bool {
	return false
}

func (t *Template) AddCode(core.Code) error {
	return errors.New("Failed to parse template, you might have an extra }")
}

func (t *Template) Close(*core.Context) error {
	panic("Close called on template tag")
}

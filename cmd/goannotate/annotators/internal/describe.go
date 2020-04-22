// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package internal

import (
	"strings"

	"cloudeng.io/cmdutil/structdoc"
)

// DocTagName is the struct tag used to document annotator configuration fields.
const DocTagName = "annotator"

func trimNL(text string) string {
	return strings.Trim(text, "\n")
}

func MustDescribe(t interface{}, detail string) string {
	detail = structdoc.TypeName(t) + ":\n" + trimNL(detail) + "\n"
	r, err := structdoc.Describe(t, DocTagName, detail)
	if err != nil {
		panic(err)
	}
	out := strings.Builder{}
	out.WriteString(r.Detail)
	out.WriteString(structdoc.FormatFields(2, 2, r.Fields))
	return out.String()
}

func Indent(text string, indent int) string {
	out := &strings.Builder{}
	spaces := strings.Repeat(" ", indent)
	out.WriteString(spaces)
	for _, r := range text {
		if r == '\n' {
			out.WriteRune('\n')
			out.WriteString(spaces)
			continue
		}
		out.WriteRune(r)
	}
	return strings.TrimSuffix(out.String(), spaces)
}

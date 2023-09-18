package templates

import (
	"fmt"
	"html/template"
)

func Field(value string) template.HTML {
	return template.HTML(value)
}

func DetailsButton(url string) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<a href="%s"><i class="fa fa-eye"></i></a>`, url),
	)
}

func Textarea(content string) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<textarea style="width: 100%%; height: 300px" readonly>%s</textarea>`, content),
	)
}

func Link(value string, url string) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<a href="%s">%s</a>`, url, value),
	)
}

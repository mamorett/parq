package pathrewrite

import (
	"regexp"

	"github.com/trithemius/parq/internal/config"
)

type Rewriter struct {
	rules []rule
}

type rule struct {
	regex   *regexp.Regexp
	replace string
}

func New(configs []config.Remap) (*Rewriter, error) {
	rules := make([]rule, 0, len(configs))
	for _, c := range configs {
		re, err := regexp.Compile(c.Pattern)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule{regex: re, replace: c.Replace})
	}
	return &Rewriter{rules: rules}, nil
}

func (r *Rewriter) Rewrite(path string) string {
	for _, rule := range r.rules {
		if rule.regex.MatchString(path) {
			return rule.regex.ReplaceAllString(path, rule.replace)
		}
	}
	return path
}

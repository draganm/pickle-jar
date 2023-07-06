package eval

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cucumberexpressions "github.com/cucumber/cucumber-expressions/go/v16"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/dop251/goja"
	"github.com/draganm/pickle-jar/jsfiles"
)

var parameterTypes = cucumberexpressions.NewParameterTypeRegistry()
var parameterGenerator = cucumberexpressions.NewCucumberExpressionGenerator(parameterTypes)

type testStepCallable struct {
	Matcher  string
	Expr     cucumberexpressions.Expression
	Callable goja.Callable
}

type TestEvaluator struct {
	// stepCallables []testStepCallable
	steps []scriptStep
}

func (te *TestEvaluator) Run(ctx context.Context) error {
	for _, s := range te.steps {
		args := []goja.Value{}

		for _, a := range s.args {
			args = append(args, s.rt.ToValue(a.GetValue()))
		}

		fmt.Println(s.step.Text)
		_, err := s.Callable(goja.Null(), args...)
		if err != nil {
			return fmt.Errorf("executing %q failed: %w", s.step.Text, err)
		}

	}
	return nil
}

type scriptStep struct {
	rt   *goja.Runtime
	step *messages.Step
	goja.Callable
	args []*cucumberexpressions.Argument
}

func ProvideEvaluator(path string, jsFiles *jsfiles.Dir, script []*messages.Step) (*TestEvaluator, error) {

	rt := goja.New()
	rt.SetFieldNameMapper(goja.UncapFieldNameMapper())

	stepDefinitions := []testStepCallable{}

	addStepCallable := func(matcher string, handler goja.Callable) error {
		expr, err := cucumberexpressions.NewCucumberExpression(matcher, parameterTypes)
		if err != nil {
			return fmt.Errorf("could not parse cucumber expression: %w", err)
		}

		stepDefinitions = append(stepDefinitions, testStepCallable{
			Matcher:  matcher,
			Callable: handler,
			Expr:     expr,
		})

		return nil
	}

	err := errors.Join(
		rt.GlobalObject().Set("Given", addStepCallable),
		rt.GlobalObject().Set("When", addStepCallable),
		rt.GlobalObject().Set("Then", addStepCallable),
	)

	if err != nil {
		return nil, fmt.Errorf("could not set Given/When/Then functions: %w", err)
	}

	definitions := jsFiles.AllStepDefinitions(path)
	for _, d := range definitions {
		// fmt.Println("included", d.Path)
		_, err = rt.RunScript(d.Path, d.Content)
		if err != nil {
			return nil, fmt.Errorf("could not execute %s: %w", d.Path, err)
		}

	}

	te := &TestEvaluator{}

	for _, st := range script {
		st := st
		matches := []scriptStep{}
		for _, sd := range stepDefinitions {
			args, err := sd.Expr.Match(st.Text)
			if err != nil {
				return nil, fmt.Errorf("could not match %q: %w", st.Text, err)
			}
			if args != nil {
				matches = append(matches, scriptStep{step: st, args: args, Callable: sd.Callable, rt: rt})
			}
		}

		if len(matches) == 0 {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("could not find step definition for %s%s\n", st.Keyword, st.Text))
			sb.WriteString("Please add one of the following snippets:\n\n")
			for _, e := range parameterGenerator.GenerateExpressions(st.Text) {

				sb.WriteString(strings.TrimSpace(st.Keyword))
				sb.WriteString(fmt.Sprintf("(%q, (%s) => {\n\n})", e.Source(), strings.Join(e.ParameterNames(), ", ")))
				sb.WriteString("\n\n")
			}

			return nil, fmt.Errorf(sb.String())
		}

		if len(matches) != 1 {
			return nil, fmt.Errorf("step %q is matched by %d matchers", st.Text, len(matches))
		}
		te.steps = append(te.steps, matches...)
	}

	return te, nil
}

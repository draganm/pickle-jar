package picklejar

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/draganm/pickle-jar/eval"
	"github.com/draganm/pickle-jar/jsfiles"
)

type feature struct {
	path            string
	gherkinDocument *messages.GherkinDocument
}

type testRun func(ctx context.Context) error

func (f *feature) createTestRuns(jsFiles *jsfiles.Dir) ([]testRun, error) {

	runs := []testRun{}

	if f.gherkinDocument.Feature == nil {
		return nil, nil
	}

	for _, ch := range f.gherkinDocument.Feature.Children {
		ch := ch
		if ch.Scenario != nil {
			ev, err := eval.ProvideEvaluator(f.path, jsFiles, ch.Scenario.Steps)
			if err != nil {
				return nil, err
			}
			runs = append(runs, ev.Run)
		}
	}

	return runs, nil
}

func RunTests(featuresDir string) error {

	idGenerator := (&messages.Incrementing{}).NewId

	features := []feature{}
	jsFiles := jsfiles.New()

	err := filepath.WalkDir(featuresDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".js" {
			d, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			jsFiles.AddFile(path, string(d))
		}

		if filepath.Ext(path) == ".feature" {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("could not open %s: %w", path, err)
			}
			defer f.Close()
			doc, err := gherkin.ParseGherkinDocument(f, idGenerator)
			if err != nil {
				return fmt.Errorf("could not parse %s: %w", path, err)
			}

			features = append(features, feature{
				path:            path,
				gherkinDocument: doc,
			})
		}

		return nil
	})

	for _, f := range features {
		trs, err := f.createTestRuns(jsFiles)
		if err != nil {
			return fmt.Errorf("could not create test runs: %w", err)
		}
		for _, tr := range trs {
			err = tr(context.Background())
			if err != nil {
				return err
			}
		}
	}

	if err != nil {
		return err
	}
	return nil
}

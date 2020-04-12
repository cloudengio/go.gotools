package annotators

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"cloudeng.io/errors"
	"cloudeng.io/text/edit"
)

func applyEdits(ctx context.Context, edits map[string][]edit.Delta) error {
	errs := &errors.M{}
	for file, edits := range edits {
		if len(edits) > 0 {
			fmt.Printf("%v\n", file)
		}
		for _, edit := range edits {
			Verbosef("\t%s: %s: %.30s...\n", file, edit, edit.Text())
		}
		if err := editFile(ctx, file, edits); err != nil {
			errs.Append(fmt.Errorf("failed to edit file: %v: %v", file, err))
		}
	}
	if err := errs.Err(); err != nil {
		return fmt.Errorf("failed to edit files: %v\n", err)
	}
	return nil
}

func editFile(ctx context.Context, name string, deltas []edit.Delta) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	buf = edit.Do(buf, deltas...)
	cmd := exec.CommandContext(ctx, "goimports")
	cmd.Stdin = bytes.NewBuffer(buf)
	out, err := cmd.Output()
	if err != nil {
		// This is most likely because the edit messed up the go code
		// and goimports is unhappy with it as its input. To help with
		// debugging write the edited code to a temp file.
		if Verbose {
			if tmpfile, err := ioutil.TempFile("", "annotate-"); err == nil {
				io.Copy(tmpfile, bytes.NewBuffer(buf))
				tmpfile.Close()
				fmt.Printf("wrote modified contents of %v to %v\n", name, tmpfile.Name())
			}
		}
		return fmt.Errorf("%v: %v", strings.Join(cmd.Args, " "), err)
	}
	return ioutil.WriteFile(name, out, info.Mode().Perm())
}

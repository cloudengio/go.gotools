// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cloudeng.io/errors"
	"cloudeng.io/path/cloudpath"
	"cloudeng.io/text/edit"
)

func computeOutputs(writeDir string, edits map[string][]edit.Delta) map[string]string {
	outputs := map[string]string{}
	if len(writeDir) == 0 || len(edits) == 0 {
		for k := range edits {
			outputs[k] = k
		}
		return outputs
	}
	var prefix string
	if len(edits) > 1 {
		filepaths := make([]cloudpath.T, 0, len(edits))
		for k := range edits {
			filepaths = append(filepaths, cloudpath.Split(k, filepath.Separator))
		}
		lcp := cloudpath.LongestCommonPrefix(filepaths)
		prefix = lcp.Join(filepath.Separator)
	} else {
		// no common prefix, so just use the basename of the supplied file.
		for k := range edits {
			prefix = filepath.Dir(k)
			break
		}
	}
	for k := range edits {
		outputs[k] = filepath.Join(writeDir, strings.TrimPrefix(k, prefix))
	}
	return outputs
}

func applyEdits(ctx context.Context, outputs map[string]string, edits map[string][]edit.Delta) error {
	errs := &errors.M{}
	for file, edits := range edits {
		fmt.Println(file)
		for _, edit := range edits {
			Verbosef("\t%s: %s: %.30s...\n", file, edit, edit.Text())
		}
		output := outputs[file]
		if len(output) == 0 {
			output = file
		} else {
			dir := filepath.Dir(output)
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("failed to create dir %v ", dir)
			}
		}
		if err := editFile(ctx, file, output, edits); err != nil {
			errs.Append(fmt.Errorf("failed to edit file: %v: %v", file, err))
		}
	}
	return errs.Err()
}

func editFile(ctx context.Context, src, dst string, deltas []edit.Delta) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	buf = edit.Do(buf, deltas...)
	cmd := exec.CommandContext(ctx, "goimports")
	cmd.Stdin = bytes.NewBuffer(buf)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if execerr, ok := err.(*exec.ExitError); ok {
			stderr = string(execerr.Stderr)
		}
		// This is most likely because the edit messed up the go code
		// and goimports is unhappy with it as its input. To help with
		// debugging write the edited code to a temp file.
		if tmpfile, err := os.CreateTemp("", "annotate-"); err == nil {
			_, _ = io.Copy(tmpfile, bytes.NewBuffer(buf))
			tmpfile.Close()
			fmt.Printf("wrote modified contents of %v to %v\n", src, tmpfile.Name())
			if len(stderr) > 0 {
				fmt.Println(stderr)
			}
		}
		return fmt.Errorf("%v: %v", strings.Join(cmd.Args, " "), err)
	}
	return os.WriteFile(dst, out, info.Mode().Perm())
}

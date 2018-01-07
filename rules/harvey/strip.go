// Copyright 2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harvey

import (
	"fmt"
	"io"
	"path/filepath"

	"bldy.build/build/executor"
	"bldy.build/build/project"
	"bldy.build/build/racy"
)

type Strip struct {
	Name         string   `strip:"name"`
	Dependencies []string `strip:"deps"`
}

func (s *Strip) GetName() string {
	return s.Name
}

func (s *Strip) GetDependencies() []string {
	return s.Dependencies
}

func (s *Strip) Hash() []byte {
	h := racy.New()
	io.WriteString(h, s.Name)
	return []byte{}
}

// Had to be done
func Stripper() string {
	if tpfx := project.Getenv("TOOLPREFIX"); tpfx == "" {
		return "strip"
	} else {
		return fmt.Sprintf("%s%s", tpfx, "strip")
	}
}
func (s *Strip) Build(e *executor.Executor) error {
	params := []string{"-o"}
	params = append(params, s.Name)
	params = append(params, filepath.Join("bin", split(s.Dependencies[0], ":")))
	if err := e.Exec(Stripper(), nil, params); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
func (s *Strip) Installs() map[string]string {
	installs := make(map[string]string)
	installs[filepath.Join("bin", s.Name)] = s.Name
	return installs
}
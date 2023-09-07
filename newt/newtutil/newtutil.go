/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package newtutil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"mynewt.apache.org/newt/newt/interfaces"
	"mynewt.apache.org/newt/util"
)

var NewtVersion = Version{1, 11, 9900}
var NewtVersionStr = "1.12.0-dev"
var NewtGitHash = "unknown"
var NewtDate = "unknown"

var NewtBlinkyTag string = "master"
var NewtNumJobs int
var NewtForce bool
var NewtAsk bool

const CORE_REPO_NAME string = "apache-mynewt-core"
const ARDUINO_ZERO_REPO_NAME string = "mynewt_arduino_zero"

type Version struct {
	Major    int64
	Minor    int64
	Revision int64
}

func ParseVersion(s string) (Version, error) {
	v := Version{}
	parseErr := util.FmtNewtError("Invalid version string: %s", s)

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return v, parseErr
	}

	var err error

	v.Major, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return v, parseErr
	}

	v.Minor, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return v, parseErr
	}

	v.Revision, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return v, parseErr
	}

	return v, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Revision)
}

func VerCmp(v1 Version, v2 Version) int64 {
	if r := v1.Major - v2.Major; r != 0 {
		return r
	}

	if r := v1.Minor - v2.Minor; r != 0 {
		return r
	}

	if r := v1.Revision - v2.Revision; r != 0 {
		return r
	}

	return 0
}

// Parses a string of the following form:
//     [@repo]<path/to/package>
//
// @return string               repo name ("" if no repo)
//         string               package name
//         error                if invalid package string
func ParsePackageString(pkgStr string) (string, string, error) {
	// remove possible trailing '/'
	pkgStr = strings.TrimSuffix(pkgStr, "/")

	if strings.HasPrefix(pkgStr, "@") {
		nameParts := strings.SplitN(pkgStr[1:], "/", 2)
		if len(nameParts) == 1 {
			return "", "", util.NewNewtError(fmt.Sprintf("Invalid package "+
				"string; contains repo but no package name: %s", pkgStr))
		} else {
			return nameParts[0], nameParts[1], nil
		}
	} else {
		return "", pkgStr, nil
	}
}

func FindRepoDesignator(s string) (int, int) {
	start := strings.Index(s, "@")
	if start == -1 {
		return -1, -1
	}

	len := strings.Index(s[start:], "/")
	if len == -1 {
		return -1, -1
	}

	return start, len
}

func ReplaceRepoDesignators(s string) (string, bool) {
	start, len := FindRepoDesignator(s)
	if start == -1 {
		return s, false
	}
	repoName := s[start+1 : start+len]

	proj := interfaces.GetProject()
	repoPath := proj.FindRepoPath(repoName)
	if repoPath == "" {
		return s, false
	}

	// Trim common project base from repo path.
	relRepoPath := strings.TrimPrefix(repoPath, proj.Path()+"/")

	return s[:start] + relRepoPath + s[start+len:], true
}

func BuildPackageString(repoName string, pkgName string) string {
	if repoName != "" {
		return "@" + repoName + "/" + pkgName
	} else {
		return pkgName
	}
}

func GeneratedPreamble() string {
	return fmt.Sprintf(
		"/**\n * This file was generated by Apache newt version: %s\n */\n\n",
		NewtVersionStr)
}

// Creates a temporary directory for downloading a repo.
func MakeTempRepoDir() (string, error) {
	tmpdir, err := ioutil.TempDir("", "newt-repo")
	if err != nil {
		return "", util.ChildNewtError(err)
	}

	return tmpdir, nil
}

func ProjRelPath(path string) string {
	if filepath.IsAbs(path) {
		proj := interfaces.GetProject()
		if proj != nil {
			relPath, err := filepath.Rel(proj.Path(), path)
			if err == nil {
				path = relPath
			}
		}
	}

	return path
}

func PrintNewtVersion() {
	util.StatusMessage(util.VERBOSITY_DEFAULT, "Apache Newt %s / %s / %s\n",
		NewtVersionStr, NewtGitHash, NewtDate)
}

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	nfpm "github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	"github.com/goreleaser/nfpm/v2/files"
	_ "github.com/goreleaser/nfpm/v2/rpm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	rootCmd = &cobra.Command{
		Use:   "make",
		Short: "purego makefile",
		Long: `The build, release, and distro tool for this project. Compile and run it on the fly with 'go run make.go'.
This depends on the availability of...

   gh: github command line tool
	 docker: container management
	 dpkg: debian package development
		`,
	}
)

func init() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName(".semver")
	_ = viper.ReadInConfig()

	rootCmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "setup for build and release",
		Long:  "pulls docker images and manages dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "basic build",
		Long:  "a simple wrapper for `go build -o build/main main.go`",
		RunE: func(cmd *cobra.Command, args []string) error {
			return build()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "buildx",
		Short: "cross platform build",
		Long:  "runs cross platform builds using a docker image configured specifically for this",
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildx()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "package",
		Short: "packages artifacts for target distros",
		Long:  "uses nfpm to generate packages for different distros",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pack()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "release",
		Short: "releases artifacts to github as a versioned release",
		Long:  "uses github cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return release()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "distro",
		Short: "distribute artifacts to distro repositories",
		Long:  "supports brew and apk for now",
		RunE: func(cmd *cobra.Command, args []string) error {
			return distro()
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setup() (err error) {
	err = ExecAndStream("docker", "pull", "jkassis/xgo:1.19.5")
	if err != nil {
		return err
	}
	return ExecAndStream("docker", "build", "-t", "jkassis/gas-builder:latest", "./bin")
}

func build() error {
	return ExecAndStream("go", "build", "-o", "dist/main", "main.go")
}

// This is complicated...
// Currently using Docker image built from github.com/jkassis/xgo
// might want to try https://github.com/crazy-max/goxx
// or https://github.com/techknowlogick/xgo
//
// Needed environment variables:
//
//	DEPS           - Optional list of C dependency packages to buildx
//	ARGS           - Optional arguments to pass to C dependency configure scripts
//	OUT            - Optional output prefix to override the package name
//	FLAG_V         - Optional verbosity flag to set on the Go builder
//	FLAG_X         - Optional flag to print the buildx progress commands
//	FLAG_RACE      - Optional race flag to set on the Go builder
//	FLAG_TAGS      - Optional tag flag to set on the Go builder
//	FLAG_LDFLAGS   - Optional ldflags flag to set on the Go builder
//	FLAG_BUILDMODE - Optional buildmode flag to set on the Go builder
//	FLAG_TRIMPATH  - Optional trimpath flag to set on the Go builder
//	TARGETS        - Comma separated list of buildx targets to compile for
//	GO_VERSION     - Bootstrapped version of Go to disable uncupported targets
//	EXT_GOPATH     - GOPATH elements mounted from the host filesystem
//
// note that cross architecture buildx with CGO dependencies can only happen
// with dedicated hardware.
//
// buildx uses docker for cross platform builds
func buildx() (err error) {
	var pwd string
	pwd, err = os.Getwd()
	if err != nil {
		return err
	}

	xgoCacheDir := os.Getenv("GOPATH") + `/xgo-cache`
	_, err = os.Stat(xgoCacheDir)
	if os.IsNotExist(err) {
		os.MkdirAll(xgoCacheDir, os.ModePerm)
	}

	err = ExecAndStream(
		"docker",
		"run", "--rm",
		"-v", pwd+"/build:/build", // build volume
		"-v", xgoCacheDir+":/deps-cache", //  xgo cache volume
		"-v", pwd+":/source", // sourcecode
		"-e", "FLAG_V=false",
		"-e", "FLAG_X=false",
		"-e", "FLAG_RACE=false ",
		"-e", "FLAG_TAGS=static",
		`-e", "FLAG_LDFLAGS="-w -s -D_REENTRANT -I/usr/include/SDL2 -lSDL2 -lSDL2main -lSDL2_image -lSDL2_mixer -lSDL2_ttf"`,
		"-e", "FLAG_BUILDMODE=default ",
		"-e", "PKG_CONFIG_PATH=/usr/local/lib/pkgconfig",
		"-e", `TARGETS="linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64`,
		"-e", `CGO_CXXFLAGS="-I/usr/include/SDL2 -D_REENTRANT"`,
		"-e", `CGO_CPPFLAGS="-I/usr/include/SDL2 -D_REENTRANT"`,
		"-e", "OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1",
		"jkassis/gas-builder:latest",
		"main.go")
	if err != nil {
		return fmt.Errorf("docker error: %v", err)
	}

	err = ExecAndStream(
		"docker",
		"run", "--rm",
		"-v", pwd+"/build:/build", // build volume
		"-v", pwd+"/bin:/opt/app", // build script
		"-v", pwd+":/source", // sourcecode
		"jkassis/gas-builder:latest")
	if err != nil {
		return fmt.Errorf("docker error: %v", err)
	}

	// change permissions of all executables
	return filepath.WalkDir(pwd+"/build", func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(path, 0555)
	})
}

func pack() (err error) {
	type Job [4]string

	// clean up the dist directory
	fmt.Printf("cleaning dist dir\n")
	filepath.WalkDir("dist", func(fp string, dirEntry os.DirEntry, err error) error {
		if err != nil || fp == "dist" {
			return err
		}
		return os.Remove(fp)
	})

	doOne := func(job Job) error {
		packager := job[0]
		platform := job[1]
		arch := job[2]
		target := job[3]

		fmt.Printf("using %s packager...\n", packager)

		pkg, err := nfpm.Get(packager)
		if err != nil {
			return err
		}

		config := &nfpm.Config{
			Info: nfpm.Info{
				Name:          "gitall",
				Arch:          arch,
				Platform:      platform,
				Version:       "1.0.1",
				VersionSchema: "semver",
				Section:       "default",
				Priority:      "extra",
				Maintainer:    "Jeremy Kassis<jkassis@gmail.com>",
				Description:   "CLI to perform git operations on multiple repos at once.",
				Homepage:      "https://github.com/jkassis/gitall",
				License:       "CC0_1.0",
				Changelog:     "changelog.md",
				Overridables: nfpm.Overridables{
					Contents: files.Contents{
						&files.Content{
							Source:      "./build/github.com/jkassis/gitall-linux-" + arch,
							Destination: "/usr/bin/gitall",
						},
					},
				},
			},
		}

		// allow packager to augment config
		info, err := config.Get(packager)
		if err != nil {
			return err
		}
		info = nfpm.WithDefaults(info)

		// open the target file
		target = path.Join(target, pkg.ConventionalFileName(info))
		f, err := os.Create(target)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := pkg.Package(info, f); err != nil {
			return err
		}

		return nil
	}

	jobs := []Job{
		{"archlinux", "linux", "amd64", "./dist"},
		{"archlinux", "linux", "arm64", "./dist"},
		{"apk", "linux", "amd64", "./dist"},
		{"apk", "linux", "arm64", "./dist"},
		{"deb", "linux", "amd64", "./dist"},
		{"deb", "linux", "arm64", "./dist"},
		{"rpm", "linux", "amd64", "./dist"},
		{"rpm", "linux", "arm64", "./dist"},
		{"rpm", "darwin", "amd64", "./dist"},
		{"rpm", "darwin", "arm64", "./dist"},
	}

	for _, job := range jobs {
		err := doOne(job)
		if err != nil {
			return err
		}
	}

	return nil
}

func release() error {
	// prompt the user to login to gh as needed
	_, authStatusErr, err := Exec("gh", "auth", "status")
	if err != nil {
		if strings.Contains(authStatusErr, "not logged into") {
			err = ExecAndStream("gh", "auth", "login")
			if err != nil {
				return fmt.Errorf("error connecting with github: %v", err)
			}
		} else {
			return fmt.Errorf("gh auth status error: %v", err)
		}
	}

	// check for local changes
	fmt.Printf("git: checking for changes\n")
	gitStatusOut, _, err := Exec("git", "status", "--porcelain")
	if err != nil || len(gitStatusOut) > 0 {
		return fmt.Errorf("git: there are uncommitted changes to this repo. Commit changes and build with bin/build.sh first: %v", err)
	}
	fmt.Printf("git: no changes\n")

	// check that branches are in sync
	fmt.Printf("git: checking that local branch is in sync with origin\n")
	branch, _, err := Exec("git", "rev-parse", "--abbrev-ref", "HEAD")
	branch = strings.TrimSpace(branch)
	if err != nil {
		return fmt.Errorf("git: problem getting branch info: %v", err)
	}

	localRev, _, err := Exec("git", "rev-parse", branch)
	localRev = strings.TrimSpace(localRev)
	if err != nil {
		return fmt.Errorf("git: problem getting local branch revision: %v", err)
	}

	remoteRev, _, err := Exec("git", "rev-parse", "origin/"+branch)
	remoteRev = strings.TrimSpace(remoteRev)
	if err != nil {
		return fmt.Errorf("git: problem getting remote branch revision: %v", err)
	}

	if localRev != remoteRev {
		return fmt.Errorf("git: %s is not in sync with origin/%s. You need to rebase or push first", branch, branch)
	}

	fmt.Printf("git: branches in sync\n")

	// get list of files
	var files []string
	files = make([]string, 0)
	filepath.WalkDir("dist", func(fp string, _ os.DirEntry, err error) error {
		if err != nil || fp == "dist" {
			return err
		}
		files = append(files, fp)
		return nil
	})

	// bump the patch version
	v, err := semver.NewVersion(viper.GetString("release"))
	if err != nil {
		return err
	}

	*v = v.IncPatch()

	fmt.Printf("bumping .semver.yaml file to %s\n", v.String())
	viper.Set("release", v)
	err = viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("could not write semver.yaml: %v", err)
	}

	// commit
	fmt.Printf("adding .semver.yaml for a new commit\n")
	err = ExecAndStream("git", "add", ".semver.yaml")
	if err != nil {
		return fmt.Errorf("trouble adding: %v", err)
	}

	// commit
	fmt.Printf("commiting\n")
	err = ExecAndStream("git", "commit", "-m", ".semver.yaml bump")
	if err != nil {
		return fmt.Errorf("trouble commiting: %v", err)
	}

	// push
	fmt.Printf("pushing%v\n", v.String())
	err = ExecAndStream("git", "push")
	if err != nil {
		return fmt.Errorf("trouble pushing: %v", err)
	}

	// tag the release
	fmt.Printf("tagging the release with %s\n", v.String())
	err = ExecAndStream("git", "tag", v.String())
	if err != nil {
		return fmt.Errorf("trouble tagging the release with git: %v", err)
	}

	// push tags
	fmt.Printf("pushing tags %v\n", v.String())
	err = ExecAndStream("git", "push", "--tags")
	if err != nil {
		return fmt.Errorf("trouble tagging the release with git: %v", err)
	}

	// create the github release
	fmt.Printf("creating the github release\n")
	err = ExecAndStream("gh", append([]string{"release", "create", v.String()}, files...)...)
	if err != nil {
		return fmt.Errorf("trouble creating the github release: %v", err)
	}
	return nil
}

func distro() error {
	return nil
}

func ExecAndStream(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Exec(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	errR, errW := io.Pipe()
	cmd.Stderr = errW
	outR, outW := io.Pipe()
	cmd.Stdout = outW

	var stdout, stderr []byte
	eg := new(errgroup.Group)
	eg.Go(func() (err error) { stdout, err = io.ReadAll(outR); return })
	eg.Go(func() (err error) { stderr, err = io.ReadAll(errR); return })
	eg.Go(func() (err error) { err = cmd.Run(); outW.Close(); errW.Close(); return })
	return string(stdout), string(stderr), eg.Wait()
}

func CP(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile("./dist"+path.Base(dst), input, 0644)
}

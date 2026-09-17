package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"setfta/internal/assoc"
	"setfta/internal/userchoice"
)

const version = "1.0.0"

var (
	progIDPattern    = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	extensionPattern = regexp.MustCompile(`^\.[A-Za-z0-9]+$`)
)

type appFlags []assoc.App

func (values *appFlags) String() string { return fmt.Sprintf("%d app(s)", len(*values)) }

func (values *appFlags) Set(value string) error {
	parts := strings.Split(value, "|")
	if len(parts) != 4 {
		return errors.New(`--app must be "ProgId|Display name|Executable|.ext,.ext"`)
	}
	exts := strings.Split(parts[3], ",")
	for i := range exts {
		exts[i] = strings.TrimSpace(exts[i])
	}
	*values = append(*values, assoc.App{
		ProgID:     strings.TrimSpace(parts[0]),
		Name:       strings.TrimSpace(parts[1]),
		Executable: strings.TrimSpace(parts[2]),
		Extensions: exts,
	})
	return nil
}

func main() {
	os.Exit(run())
}

func run() int {
	var apps appFlags
	yes := flag.Bool("yes", false, "apply without waiting for Enter")
	noPause := flag.Bool("no-pause", false, "do not wait for Enter before exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Var(&apps, "app", `repeatable: "ProgId|Display name|Executable|.ext,.ext"`)
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println("SetFTA", version)
		return 0
	}
	if flag.NArg() != 0 || len(apps) == 0 {
		usage()
		return 2
	}
	if err := validate(apps); err != nil {
		fmt.Fprintln(os.Stderr, "Configuration error:", err)
		pauseUnless(*noPause)
		return 2
	}
	if !assoc.IsElevated() {
		fmt.Fprintln(os.Stderr, "Administrator rights are required. Run the caller as administrator.")
		pauseUnless(*noPause)
		return 3
	}
	if got := userchoice.Hash(".txt", "s-1-5-21-100", "portableassoc.editor", "01dc000000000000"); got != "UBxHti9ONDw=" {
		fmt.Fprintln(os.Stderr, "Internal hash self-test failed; no changes were made.")
		pauseUnless(*noPause)
		return 4
	}

	printPreview(apps)
	if !*yes {
		waitForEnter("Press Enter to apply, or Ctrl+C to cancel: ")
	}

	tempDir, err := os.MkdirTemp("", "setfta-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Create temporary directory:", err)
		pauseUnless(*noPause)
		return 1
	}
	defer os.RemoveAll(tempDir)

	succeeded, failed := 0, 0
	for _, app := range apps {
		if err := assoc.RegisterApp(app); err != nil {
			fmt.Printf("[FAIL] %s: register ProgId: %v\n", app.Name, err)
			failed += len(app.Extensions)
			continue
		}
		for _, extension := range app.Extensions {
			if err := assoc.SetUserChoice(extension, app.ProgID, tempDir); err != nil {
				fmt.Printf("[FAIL] %s -> %v\n", extension, err)
				failed++
				continue
			}
			fmt.Printf("[OK]   %s -> %s\n", extension, app.ProgID)
			succeeded++
		}
	}
	assoc.NotifyShell()
	fmt.Printf("\nCompleted: %d succeeded, %d failed.\n", succeeded, failed)
	pauseUnless(*noPause)
	if failed != 0 {
		return 1
	}
	return 0
}

func validate(apps []assoc.App) error {
	seen := make(map[string]string)
	for i := range apps {
		app := &apps[i]
		if !progIDPattern.MatchString(app.ProgID) {
			return fmt.Errorf("unsafe ProgId %q", app.ProgID)
		}
		if app.Name == "" || strings.ContainsAny(app.Name, "\r\n|") {
			return fmt.Errorf("unsafe display name %q", app.Name)
		}
		absolute, err := filepath.Abs(app.Executable)
		if err != nil {
			return fmt.Errorf("invalid executable %q: %w", app.Executable, err)
		}
		info, err := os.Stat(absolute)
		if err != nil || info.IsDir() {
			return fmt.Errorf("executable not found: %s", absolute)
		}
		app.Executable = absolute
		for j, extension := range app.Extensions {
			extension = strings.ToLower(extension)
			if !extensionPattern.MatchString(extension) {
				return fmt.Errorf("unsafe extension %q", extension)
			}
			if owner, exists := seen[extension]; exists {
				return fmt.Errorf("duplicate extension %s in %s and %s", extension, owner, app.Name)
			}
			seen[extension] = app.Name
			app.Extensions[j] = extension
		}
	}
	return nil
}

func printPreview(apps []assoc.App) {
	fmt.Println("\nFile association preview")
	fmt.Println(strings.Repeat("=", 70))
	for _, app := range apps {
		exts := append([]string(nil), app.Extensions...)
		sort.Strings(exts)
		fmt.Printf("\n[%s]\nProgram:    %s\nProgId:     %s\nExtensions: %s\n", app.Name, app.Executable, app.ProgID, strings.Join(exts, " "))
	}
	fmt.Println("\nThis changes defaults for the current user only.")
}

func waitForEnter(prompt string) {
	fmt.Print(prompt)
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func pauseUnless(noPause bool) {
	if !noPause {
		waitForEnter("Press Enter to close this window: ")
	}
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `SetFTA %s - generic portable file association tool

Usage:
  SetFTA.exe --app "ProgId|Display name|C:\Path\App.exe|.ext,.ext" [--app ...]

Options:
`, version)
	flag.PrintDefaults()
}

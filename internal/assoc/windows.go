package assoc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"setfta/internal/userchoice"
)

type App struct {
	ProgID     string
	Name       string
	Executable string
	Extensions []string
}

func IsElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

func CurrentSID() (string, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("get current user token: %w", err)
	}
	return user.User.Sid.String(), nil
}

// RegisterApp creates a per-user ProgId for a portable executable.
func RegisterApp(app App) error {
	root := `Software\Classes\` + app.ProgID
	if err := setDefaultValue(root, app.Name); err != nil {
		return err
	}
	if err := setDefaultValue(root+`\DefaultIcon`, fmt.Sprintf(`"%s",0`, app.Executable)); err != nil {
		return err
	}
	command := fmt.Sprintf(`"%s" "%%1"`, app.Executable)
	return setDefaultValue(root+`\shell\open\command`, command)
}

func setDefaultValue(path, value string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("create registry key %s: %w", path, err)
	}
	defer key.Close()
	if err := key.SetStringValue("", value); err != nil {
		return fmt.Errorf("set registry key %s: %w", path, err)
	}
	return nil
}

// SetUserChoice updates one protected UserChoice key and verifies the result.
func SetUserChoice(extension, progID, tempDir string) error {
	sid, err := CurrentSID()
	if err != nil {
		return err
	}
	timestamp := minuteFileTime(time.Now())
	hash := userchoice.Hash(extension, sid, progID, timestamp)
	if hash == "" {
		return errors.New("hash calculation returned an empty value")
	}

	keyPath := `Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\` + extension + `\UserChoice`
	existed := registryKeyExists(keyPath)
	regPath := `\Registry\User\` + sid + `\Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\` + extension + `\UserChoice`
	var input strings.Builder
	if existed {
		fmt.Fprintf(&input, "%s [DELETE]\r\n", regPath)
	}
	fmt.Fprintf(&input, "%s\r\n    ProgId = REG_SZ %s\r\n    Hash = REG_SZ %s\r\n", regPath, progID, hash)

	iniPath := filepath.Join(tempDir, strings.TrimPrefix(extension, ".")+".ini")
	if err := os.WriteFile(iniPath, []byte(input.String()), 0o600); err != nil {
		return fmt.Errorf("write temporary regini input: %w", err)
	}

	regini := filepath.Join(os.Getenv("SystemRoot"), "System32", "regini.exe")
	output, err := exec.Command(regini, iniPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("regini failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	actualProgID, actualHash, err := readUserChoice(keyPath)
	if err != nil {
		return fmt.Errorf("verify %s: %w", extension, err)
	}
	if actualProgID != progID || actualHash != hash {
		return fmt.Errorf("verification mismatch: ProgId=%q Hash=%q", actualProgID, actualHash)
	}
	return nil
}

func registryKeyExists(path string) bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	key.Close()
	return true
}

func readUserChoice(path string) (string, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return "", "", err
	}
	defer key.Close()
	progID, _, err := key.GetStringValue("ProgId")
	if err != nil {
		return "", "", err
	}
	hash, _, err := key.GetStringValue("Hash")
	return progID, hash, err
}

func minuteFileTime(now time.Time) string {
	localMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	const unixToFileTime = uint64(116444736000000000)
	fileTime := uint64(localMinute.UTC().UnixNano()/100) + unixToFileTime
	return fmt.Sprintf("%016X", fileTime)
}

func NotifyShell() {
	const shcneAssocChanged = 0x08000000
	shell32 := windows.NewLazySystemDLL("shell32.dll")
	proc := shell32.NewProc("SHChangeNotify")
	_, _, _ = proc.Call(shcneAssocChanged, 0, 0, 0)
}

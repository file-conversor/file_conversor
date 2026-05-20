// internal/deps/dependency.go

package deps

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/file-conversor/file_conversor/internal/logger"
)

type Package struct {
	Name            string
	Manager         *pkgMgr
	preInstallCmds  [][]string
	postInstallCmds [][]string
}

type Dependency struct {
	Name     string
	Binary   string
	License  license
	Homepage string
	Packages []Package
	lock     sync.Mutex
}

func (d *Dependency) install(dry_run bool) error {
	for _, depMgr := range d.Packages {
		var errGrp error = nil
		var pkgMgr = depMgr.Manager

		// check if pkg mgr is available for current OS, if not, skip to next pkg mgr
		if pkgMgr.IsAvailableForCurrOS() != nil {
			continue
		}

		// try to install pkg with pkg mgr, if error, add to errGrp and try with next pkg mgr
		if err := pkgMgr.InstallAll(dry_run, depMgr.postInstallCmds, depMgr.preInstallCmds, depMgr.Name); err != nil {
			errGrp = errors.Join(errGrp, err)
			continue
		}

		// all pkg mgrs return error, show error group for pkg
		return errGrp
	}
	// no pkg mgr found for any pkg, return error
	return fmt.Errorf("no available pkg mgr found to install dependency '%s'", d.Name)
}

func (d *Dependency) promptUser() error {
	logger.Infof("This feature requires '%s' to work properly, but '%s' is NOT installed.\n", d.Name, d.Name)
	logger.Infof("\n")
	logger.Infof("'%s' is available at %s , and it is licensed under %s. \n", d.Name, d.Homepage, d.License.String())
	logger.Infof("\n")
	logger.Infof("This tool can install '%s' for the current user. In that case, the following commands will be executed: \n", d.Name)
	d.install(true)
	logger.Infof("\n")
	logger.Infof("Do you accept '%s' license, and me to continue with the installation? (y/N): ", d.Name)
	var answer string
	if _, err := fmt.Scan(&answer); err != nil {
		return fmt.Errorf("reading user input: %v", err)
	}
	answer = strings.ToUpper(answer)
	if answer[0] != 'Y' {
		return fmt.Errorf("user did not accept '%s' license (%s)", d.Name, d.License.ShortName)
	}
	return nil
}

func (d *Dependency) EnsureInstalled(prompt_user bool, dry_run bool) (string, error) {
	// Lock to prevent multiple concurrent installations of the same dependency
	d.lock.Lock()
	defer d.lock.Unlock()

	// check if binary is already available, if not, try to install it and check again
	bin, err := exec.LookPath(d.Binary)
	if err != nil {
		// binary not found, prompt user to install dependency
		if prompt_user {
			if err := d.promptUser(); err != nil {
				return "", err
			}
			logger.Warnf("User accepted the installation of '%s', licensed under '%s' ...\n", d.Name, d.License.ShortName)
		} else {
			logger.Warnf("Auto-install flag is set. User automatically accepts the installation of '%s', licensed under '%s' ...\n", d.Name, d.License.ShortName)
		}

		// try to install dependency
		logger.Infof("Installing '%s' dependency ...\n", d.Name)
		if err := d.install(dry_run); err != nil {
			return "", fmt.Errorf("dependency install '%s': %v", d.Name, err)
		}

		// check if binary is available after installation
		logger.Infof("Checking if '%s' is available ...\n", d.Name)
		bin, err := exec.LookPath(d.Binary)
		if err != nil {
			return "", fmt.Errorf("cannot find '%s' binary: %v", d.Name, err)
		}
		return bin, nil
	}
	return bin, nil
}

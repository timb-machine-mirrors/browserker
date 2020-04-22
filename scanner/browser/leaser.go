package browser

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type LeaserService interface {
	Acquire() (string, error) // returns port number
	Return(port string) error
	Cleanup() (string, error)
	Count() (string, error)
}

func randPort() string {
	l, err := net.Listen("tcp", ":0")

	if err != nil {
		log.Warn().Err(err).Msg("unable to get port using default 9022")
		return "9022"
	}
	_, randPort, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return randPort
}

func randProfile(tmp string) string {
	profile, err := ioutil.TempDir(tmp, "gcd")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary profile directory")
		return "tmp"
	}

	return profile
}

func RemoveTmpContents(tmp string) error {
	files, err := filepath.Glob(filepath.Join(tmp, "gcd*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func KillOldProcesses() error {
	cmd := exec.Command("killall", "google-chrome")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn().Msgf("google-chrome %s:%s", err.Error(), string(output))
	}

	cmd = exec.Command("killall", "chrome")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Warn().Msgf("chrome %s:%s", err.Error(), string(output))
	}
	return nil
}
